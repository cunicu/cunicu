package epdisc

import (
	"fmt"

	"errors"

	"github.com/pion/ice/v2"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/feat/epdisc/proxy"
	"go.uber.org/zap"

	errorsx "github.com/stv0g/cunicu/pkg/errors"
	protoepdisc "github.com/stv0g/cunicu/pkg/proto/feat/epdisc"
)

type Interface struct {
	*core.Interface

	Discovery *EndpointDiscovery

	nat *proxy.NAT

	natRule      *proxy.NATRule
	natRuleSrflx *proxy.NATRule

	udpMux      ice.UDPMux
	udpMuxSrflx ice.UniversalUDPMux

	udpMuxPort      int
	udpMuxSrflxPort int

	logger *zap.Logger
}

func NewInterface(ci *core.Interface, d *EndpointDiscovery) (*Interface, error) {
	i := &Interface{
		Interface: ci,
		Discovery: d,

		logger: zap.L().Named("epdisc.interface"),
	}

	// Create per-interface UDPMux
	var err error

	if i.udpMux, i.udpMuxPort, err = proxy.CreateUDPMux(); err != nil && !errors.Is(err, errorsx.ErrNotSupported) {
		return nil, fmt.Errorf("failed to setup host UDP mux: %w", err)
	}

	if i.udpMuxSrflx, i.udpMuxSrflxPort, err = proxy.CreateUniversalUDPMux(); err != nil && !errors.Is(err, errorsx.ErrNotSupported) {
		return nil, fmt.Errorf("failed to setup srflx UDP mux: %w", err)
	}

	i.logger.Info("Created UDP muxes",
		zap.Int("port-host", i.udpMuxPort),
		zap.Int("port-srflx", i.udpMuxSrflxPort))

	// Setup Netfilter PAT for non-userspace devices
	if _, ok := i.KernelDevice.(*device.UserDevice); !ok {
		// Setup NAT
		ident := fmt.Sprintf("wice-if%d", i.KernelDevice.Index())
		if i.nat, err = proxy.NewNAT(ident); err != nil && !errors.Is(err, errorsx.ErrNotSupported) {
			return nil, fmt.Errorf("failed to setup NAT: %w", err)
		}

		// Setup DNAT redirects (STUN ports -> WireGuard listen ports)
		if err := i.SetupRedirects(); err != nil {
			return nil, fmt.Errorf("failed to setup redirects: %w", err)
		}
	}

	return i, nil
}

func (i *Interface) Close() error {
	if i.nat != nil {
		if err := i.nat.Close(); err != nil {
			return fmt.Errorf("failed to de-initialize NAT: %w", err)
		}
	}

	if err := i.udpMux.Close(); err != nil {
		return fmt.Errorf("failed to do-initialize UDP mux: %w", err)
	}

	if err := i.udpMuxSrflx.Close(); err != nil {
		return fmt.Errorf("failed to do-initialize srflx UDP mux: %w", err)
	}

	return nil
}

func (i *Interface) Marshal() *protoepdisc.Interface {
	is := &protoepdisc.Interface{
		MuxPort:      uint32(i.udpMuxPort),
		MuxSrflxPort: uint32(i.udpMuxSrflxPort),
	}

	if i.nat != nil {
		is.NatType = protoepdisc.NATType_NAT_NFTABLES
	}

	return is
}

func (i *Interface) UpdateRedirects() error {
	// Userspace devices need no redirects
	if i.nat == nil {
		return nil
	}

	// Delete old rules if presetn
	if i.natRule != nil {
		if err := i.natRule.Delete(); err != nil {
			return fmt.Errorf("failed to delete rule: %w", err)
		}
	}

	if i.natRuleSrflx != nil {
		if err := i.natRuleSrflx.Delete(); err != nil {
			return fmt.Errorf("failed to delete rule: %w", err)
		}
	}

	return i.SetupRedirects()
}

func (i *Interface) SetupRedirects() error {
	var err error

	// Redirect non-STUN traffic directed at UDP muxes to WireGuard interface via in-kernel port redirect / NAT
	if i.natRule, err = i.nat.RedirectNonSTUN(i.udpMuxPort, i.ListenPort); err != nil {
		return fmt.Errorf("failed to setup port redirect for server reflexive UDP mux: %w", err)
	}

	if i.natRuleSrflx, err = i.nat.RedirectNonSTUN(i.udpMuxSrflxPort, i.ListenPort); err != nil {
		return fmt.Errorf("failed to setup port redirect for server reflexive UDP mux: %w", err)
	}

	return nil
}
