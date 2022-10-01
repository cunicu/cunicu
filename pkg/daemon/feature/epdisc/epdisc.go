// Package epdisc implements endpoint (EP) discovery using Interactive Connection Establishment (ICE).
package epdisc

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/ice/v2"
	"github.com/pion/stun"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc/proxy"
	"github.com/stv0g/cunicu/pkg/device"

	errorsx "github.com/stv0g/cunicu/pkg/errors"
	icex "github.com/stv0g/cunicu/pkg/ice"

	protoepdisc "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

func init() {
	daemon.Features["epdisc"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Endpoint discovery",
		Order:       50,
	}
}

type Interface struct {
	*daemon.Interface

	nat *proxy.NAT

	natRule      *proxy.NATRule
	natRuleSrflx *proxy.NATRule

	udpMux      ice.UDPMux
	udpMuxSrflx ice.UniversalUDPMux

	udpMuxPort      int
	udpMuxSrflxPort int

	Peers map[*core.Peer]*Peer

	onConnectionStateChange []OnConnectionStateHandler

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.EndpointDisc.Enabled {
		return nil, nil
	}

	e := &Interface{
		Interface: i,
		Peers:     map[*core.Peer]*Peer{},

		onConnectionStateChange: []OnConnectionStateHandler{},

		logger: zap.L().Named("epdisc").With(zap.String("intf", i.Name())),
	}

	// Create per-interface UDPMux
	var err error

	if e.udpMux, e.udpMuxPort, err = proxy.CreateUDPMux(); err != nil && !errors.Is(err, errorsx.ErrNotSupported) {
		return nil, fmt.Errorf("failed to setup host UDP mux: %w", err)
	}

	if e.udpMuxSrflx, e.udpMuxSrflxPort, err = proxy.CreateUniversalUDPMux(); err != nil && !errors.Is(err, errorsx.ErrNotSupported) {
		return nil, fmt.Errorf("failed to setup srflx UDP mux: %w", err)
	}

	e.logger.Info("Created UDP muxes",
		zap.Int("port-host", e.udpMuxPort),
		zap.Int("port-srflx", e.udpMuxSrflxPort))

	// Setup Netfilter PAT for non-userspace devices
	if _, ok := i.KernelDevice.(*device.UserDevice); !ok {
		// Setup NAT
		ident := fmt.Sprintf("cunicu-if%d", i.KernelDevice.Index())
		if e.nat, err = proxy.NewNAT(ident); err != nil && !errors.Is(err, errorsx.ErrNotSupported) {
			return nil, fmt.Errorf("failed to setup NAT: %w", err)
		}

		// Setup DNAT redirects (STUN ports -> WireGuard listen ports)
		if err := e.SetupRedirects(); err != nil {
			return nil, fmt.Errorf("failed to setup redirects: %w", err)
		}
	}

	i.OnModified(e)
	i.OnPeer(e)

	return e, nil
}

func (e *Interface) Start() error {
	e.logger.Info("Started endpoint discovery")

	return nil
}

func (e *Interface) Close() error {
	// First switch all sessions to closing so they do not get restarted
	for _, p := range e.Peers {
		p.setConnectionState(icex.ConnectionStateClosing)
	}

	for _, p := range e.Peers {
		if err := p.Close(); err != nil {
			return fmt.Errorf("failed to close peer: %w", err)
		}
	}

	if e.nat != nil {
		if err := e.nat.Close(); err != nil {
			return fmt.Errorf("failed to de-initialize NAT: %w", err)
		}
	}

	if err := e.udpMux.Close(); err != nil {
		return fmt.Errorf("failed to do-initialize UDP mux: %w", err)
	}

	if err := e.udpMuxSrflx.Close(); err != nil {
		return fmt.Errorf("failed to do-initialize srflx UDP mux: %w", err)
	}

	return nil
}

func (e *Interface) Marshal() *protoepdisc.Interface {
	is := &protoepdisc.Interface{
		MuxPort:      uint32(e.udpMuxPort),
		MuxSrflxPort: uint32(e.udpMuxSrflxPort),
	}

	if e.nat != nil {
		is.NatType = protoepdisc.NATType_NAT_NFTABLES
	}

	return is
}

func (e *Interface) UpdateRedirects() error {
	// Userspace devices need no redirects
	if e.nat == nil {
		return nil
	}

	// Delete old rules if present
	if e.natRule != nil {
		if err := e.natRule.Delete(); err != nil {
			return fmt.Errorf("failed to delete rule: %w", err)
		}
	}

	if e.natRuleSrflx != nil {
		if err := e.natRuleSrflx.Delete(); err != nil {
			return fmt.Errorf("failed to delete rule: %w", err)
		}
	}

	return e.SetupRedirects()
}

func (e *Interface) SetupRedirects() error {
	var err error

	// Redirect non-STUN traffic directed at UDP muxes to WireGuard interface via in-kernel port redirect / NAT
	if e.natRule, err = e.nat.RedirectNonSTUN(e.udpMuxPort, e.ListenPort); err != nil {
		return fmt.Errorf("failed to setup port redirect for server reflexive UDP mux: %w", err)
	}

	if e.natRuleSrflx, err = e.nat.RedirectNonSTUN(e.udpMuxSrflxPort, e.ListenPort); err != nil {
		return fmt.Errorf("failed to setup port redirect for server reflexive UDP mux: %w", err)
	}

	return nil
}

func (e *Interface) PeerByPublicKey(pk crypto.Key) *Peer {
	if cp, ok := e.Interface.Peers[pk]; ok {
		return e.Peers[cp]
	}

	return nil
}

// Endpoint returns the best guess about our own endpoint
func (e *Interface) Endpoint() (*net.UDPAddr, error) {
	var ep *net.UDPAddr
	var bestPrio uint32 = 0

	for _, p := range e.Peers {
		cs, err := p.agent.GetLocalCandidates()
		if err != nil {
			return nil, err
		}

		for _, c := range cs {
			switch c.Type() {
			case ice.CandidateTypeHost, ice.CandidateTypeServerReflexive:
				if !c.NetworkType().IsUDP() {
					continue
				}

				if c.Priority() > bestPrio {
					bestPrio = c.Priority()
					ep = &net.UDPAddr{
						IP:   net.ParseIP(c.Address()),
						Port: c.Port(),
					}
				}
			}
		}
	}

	// No connected peers? Initiate a STUN binding request ourself..
	if ep == nil {
		c, err := stun.Dial("tcp", "stun.cunicu.li:3478")
		if err != nil {
			panic(err)
		}

		// Building binding request with random transaction id.
		message := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

		var ip net.IP

		// Sending request to STUN server, waiting for response message.
		if err := c.Do(message, func(res stun.Event) {
			if res.Error != nil {
				panic(res.Error)
			}

			// Decoding XOR-MAPPED-ADDRESS attribute from message.
			var xorAddr stun.XORMappedAddress
			if err := xorAddr.GetFrom(res.Message); err != nil {
				panic(err)
			}

			ip = xorAddr.IP
		}); err != nil {
			return nil, err
		}

		ep = &net.UDPAddr{
			IP:   ip,
			Port: e.ListenPort,
		}
	}

	return ep, nil
}
