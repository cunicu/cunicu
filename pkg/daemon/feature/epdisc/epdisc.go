// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package epdisc implements endpoint (EP) discovery using Interactive Connection Establishment (ICE).
package epdisc

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/ice/v2"
	"github.com/pion/stun"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

var Get = daemon.RegisterFeature(New, 50) //nolint:gochecknoglobals

type Interface struct {
	*daemon.Interface

	nat *NAT

	natRule      *NATRule
	natRuleSrflx *NATRule

	// muxConns is a list of UDP connections which are used by pion/ice
	// agents for muxing
	muxConns []net.PacketConn

	mux      ice.UDPMux
	muxSrflx ice.UniversalUDPMux

	muxPort      int
	muxSrflxPort int

	Peers map[*daemon.Peer]*Peer

	logger *log.Logger
}

func New(di *daemon.Interface) (*Interface, error) {
	if !di.Settings.DiscoverEndpoints {
		return nil, daemon.ErrFeatureDeactivated
	}

	i := &Interface{
		Interface: di,
		Peers:     map[*daemon.Peer]*Peer{},

		logger: log.Global.Named("epdisc").With(zap.String("intf", di.Name())),
	}

	i.AddPeerHandler(i)
	i.AddModifiedHandler(i)
	i.Bind().AddOpenHandler(i)

	// Create per-interface UDP muxes
	if i.Settings.ICE.HasCandidateType(ice.CandidateTypeHost) {
		if err := i.setupUDPMux(); err != nil && !errors.Is(err, errNotSupported) {
			return nil, fmt.Errorf("failed to setup host UDP mux: %w", err)
		}
	}

	if i.Settings.ICE.HasCandidateType(ice.CandidateTypeServerReflexive) {
		if err := i.setupUniversalUDPMux(); err != nil && !errors.Is(err, errNotSupported) {
			return nil, fmt.Errorf("failed to setup srflx UDP mux: %w", err)
		}
	}

	// Setup Netfilter port forwarding for non-userspace devices
	if i.Settings.PortForwarding && !i.IsUserspace() {
		if err := i.setupNAT(); err != nil {
			return nil, fmt.Errorf("failed to setup NAT: %w", err)
		}
	}

	return i, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started endpoint discovery")

	return nil
}

func (i *Interface) Close() error {
	i.Bind().RemoveOpenHandler(i)

	for _, p := range i.Peers {
		if err := p.Close(); err != nil {
			return fmt.Errorf("failed to close peer '%s': %w", p, err)
		}
	}

	if i.nat != nil {
		if err := i.nat.Close(); err != nil {
			return fmt.Errorf("failed to de-initialize NAT: %w", err)
		}
	}

	if i.mux != nil {
		if err := i.mux.Close(); err != nil {
			return fmt.Errorf("failed to do-initialize UDP mux: %w", err)
		}
	}

	if i.muxSrflx != nil {
		if err := i.muxSrflx.Close(); err != nil {
			return fmt.Errorf("failed to do-initialize srflx UDP mux: %w", err)
		}
	}

	return nil
}

func (i *Interface) Marshal() *epdiscproto.Interface {
	is := &epdiscproto.Interface{}

	if i.mux != nil {
		is.MuxPort = uint32(i.muxPort)
	}

	if i.muxSrflx != nil {
		is.MuxSrflxPort = uint32(i.muxSrflxPort)
	}

	if i.nat == nil {
		is.NatType = epdiscproto.NATType_NONE
	} else {
		is.NatType = epdiscproto.NATType_NFTABLES
	}

	return is
}

func (i *Interface) PeerByPublicKey(pk crypto.Key) *Peer {
	if cp, ok := i.Interface.Peers[pk]; ok {
		return i.Peers[cp]
	}

	return nil
}

// Endpoint returns the best guess about our own endpoint
func (i *Interface) Endpoint() (*net.UDPAddr, error) {
	var ep *net.UDPAddr
	var bestPrio uint32

	for _, p := range i.Peers {
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

			case ice.CandidateTypePeerReflexive, ice.CandidateTypeRelay, ice.CandidateTypeUnspecified:
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
			Port: i.ListenPort,
		}
	}

	return ep, nil
}
