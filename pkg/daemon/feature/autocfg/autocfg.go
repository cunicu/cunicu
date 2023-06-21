// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package autocfg handles initial auto-configuration of new interfaces and peers
package autocfg

import (
	"errors"
	"fmt"
	"math"
	"net"
	"syscall"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/log"
	netx "github.com/stv0g/cunicu/pkg/net"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var errMTUTooSmall = errors.New("MTU too small")

var Get = daemon.RegisterFeature(New, 10) //nolint:gochecknoglobals

type Interface struct {
	*daemon.Interface

	logger *log.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	a := &Interface{
		Interface: i,
		logger:    log.Global.Named("autocfg").With(zap.String("intf", i.Name())),
	}

	i.AddModifiedHandler(a)
	i.AddPeerHandler(a)

	return a, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started auto-configuration")

	if err := i.ConfigureWireGuard(); err != nil {
		i.logger.Error("Failed to configure WireGuard interface", zap.Error(err))
	}

	// Assign auto-generated addresses
	if sk := i.PrivateKey(); sk.IsSet() {
		pk := sk.PublicKey()
		if err := i.AddAddresses(pk); err != nil {
			i.logger.Error("Failed to add addresses", zap.Error(err))
		}
	}

	// Autodetect MTU
	// TODO: Update MTU when peers are added or their endpoints change
	if mtu := i.Settings.MTU; mtu == 0 {
		var err error
		if mtu, err = i.DetectMTU(); err != nil {
			i.logger.Error("Failed to detect MTU", zap.Error(err))
		} else {
			if err := i.SetMTU(mtu); err != nil {
				i.logger.Error("Failed to set MTU", zap.Error(err), zap.Int("mtu", i.Settings.MTU))
			}
		}
	}

	// Set link up
	if err := i.SetUp(); err != nil {
		i.logger.Error("Failed to bring link up", zap.Error(err))
	}

	return nil
}

func (i *Interface) Close() error {
	return nil
}

// ConfigureWireGuard configures the WireGuard device using the configuration provided by the user.
// Missing settings such as a private key or listen port are automatically generated/allocated.
func (i *Interface) ConfigureWireGuard() error {
	var err error

	cfg := wgtypes.Config{}

	// Private key
	if !i.PrivateKey().IsSet() || (i.Settings.PrivateKey.IsSet() && i.Settings.PrivateKey != i.PrivateKey()) {
		sk := i.Settings.PrivateKey
		if !sk.IsSet() {
			i.logger.Info("Device has no private key. Generating a new key.")

			sk, err = crypto.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("failed to generate private key: %w", err)
			}
		}

		cfg.PrivateKey = (*wgtypes.Key)(&sk)
	}

	// Listen port
	if i.ListenPort == 0 || (i.Settings.ListenPort != nil && i.ListenPort != *i.Settings.ListenPort) {
		if i.Settings.ListenPort != nil {
			cfg.ListenPort = i.Settings.ListenPort
		} else {
			port, err := netx.FindNextPortToListen("udp",
				i.Settings.ListenPortRange.Min,
				i.Settings.ListenPortRange.Max,
			)
			if err != nil {
				return fmt.Errorf("failed set listen port: %w", err)
			}

			cfg.ListenPort = &port
		}

		if i.ListenPort == 0 {
			i.logger.Info("Device has no listen port. Assigning one.", zap.Int("listen_port", *cfg.ListenPort))
		}
	}

	if cfg.PrivateKey != nil || cfg.ListenPort != nil {
		if err := i.ConfigureDevice(cfg); err != nil {
			return fmt.Errorf("failed to configure device: %w", err)
		}
	}

	return nil
}

// DetectMTU find a suitable MTU for the tunnel interface.
// The algorithm is the same as used by wg-quick:
//
//	The MTU is automatically determined from the endpoint addresses
//	or the system default route, which is usually a sane choice.
func (i *Interface) DetectMTU() (mtu int, err error) {
	mtu = math.MaxInt
	for _, p := range i.Peers {
		if p.Endpoint != nil {
			if pmtu, err := link.DetectMTU(p.Endpoint.IP, i.FirewallMark); err != nil {
				return -1, err
			} else if pmtu < mtu {
				mtu = pmtu
			}
		}
	}

	if mtu == math.MaxInt {
		if mtu, err = link.DetectDefaultMTU(i.FirewallMark); err != nil {
			return -1, err
		}
	}

	if mtu-wg.TunnelOverhead < wg.MinimalMTU {
		return -1, fmt.Errorf("%w: %d", errMTUTooSmall, mtu)
	}

	return mtu - wg.TunnelOverhead, nil
}

func (i *Interface) RemoveAddresses(pk crypto.Key) error {
	for _, pfx := range i.Settings.Prefixes {
		addr := pk.IPAddress(pfx)
		if err := i.Device.DeleteAddress(addr); err != nil {
			return err
		}

		// On Darwin systems, the utun interfaces are point-to-point
		// links which are only configured a source/destination address
		// pair. Hence we need to setup dedicated routes.
		if i.Device.Flags()&net.FlagPointToPoint != 0 {
			rte := net.IPNet{
				IP:   addr.IP.Mask(addr.Mask),
				Mask: addr.Mask,
			}

			if err := i.Device.DeleteRoute(rte, i.Settings.RoutingTable); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Interface) AddAddresses(pk crypto.Key) error {
	for _, pfx := range i.Settings.Prefixes {
		addr := pk.IPAddress(pfx)
		if err := i.Device.AddAddress(addr); err != nil && !errors.Is(err, syscall.EEXIST) {
			return err
		}

		// On Darwin systems, the utun interfaces are point-to-point
		// links which are only configured a source/destination address
		// pair. Hence we need to setup dedicated routes.
		if i.Device.Flags()&net.FlagPointToPoint != 0 {
			rte := net.IPNet{
				IP:   addr.IP.Mask(addr.Mask),
				Mask: addr.Mask,
			}

			if err := i.Device.AddRoute(rte, nil, i.Settings.RoutingTable); err != nil {
				return err
			}
		}
	}

	return nil
}
