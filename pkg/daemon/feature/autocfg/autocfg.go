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
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func init() {
	daemon.RegisterFeature("autocfg", "Auto configuration", New, 10)
}

type Interface struct {
	*daemon.Interface

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	a := &Interface{
		Interface: i,
		logger:    zap.L().Named("autocfg").With(zap.String("intf", i.Name())),
	}

	i.OnModified(a)
	i.OnPeer(a)

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
			if err := i.KernelDevice.SetMTU(mtu); err != nil {
				i.logger.Error("Failed to set MTU", zap.Error(err), zap.Int("mtu", i.Settings.MTU))
			}
		}
	}

	// Set link up
	if err := i.KernelDevice.SetUp(); err != nil {
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
			i.logger.Warn("Device has no private key. Setting a random one.")

			sk, err = crypto.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("failed to generate private key: %w", err)
			}
		}

		cfg.PrivateKey = (*wgtypes.Key)(&sk)
	}

	// Listen port
	if i.ListenPort == 0 || (i.Settings.ListenPort != nil && i.ListenPort != *i.Settings.ListenPort) {
		if i.ListenPort == 0 {
			i.logger.Warn("Device has no listen port. Setting a random one.")
		}

		if i.Settings.ListenPort != nil {
			cfg.ListenPort = i.Settings.ListenPort
		} else {
			port, err := util.FindNextPortToListen("udp",
				i.Settings.ListenPortRange.Min,
				i.Settings.ListenPortRange.Max,
			)
			if err != nil {
				return fmt.Errorf("failed set listen port: %w", err)
			}

			cfg.ListenPort = &port
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
			if pmtu, err := device.DetectMTU(p.Endpoint.IP, i.FirewallMark); err != nil {
				return -1, err
			} else if pmtu < mtu {
				mtu = pmtu
			}
		}
	}

	if mtu == math.MaxInt {
		if mtu, err = device.DetectDefaultMTU(i.FirewallMark); err != nil {
			return -1, err
		}
	}

	if mtu-wg.TunnelOverhead < wg.MinimalMTU {
		return -1, fmt.Errorf("MTU too small: %d", mtu)
	}

	return mtu - wg.TunnelOverhead, nil
}

func (i *Interface) RemoveAddresses(pk crypto.Key) error {
	for _, pfx := range i.Settings.Prefixes {
		addr := pk.IPAddress(pfx)
		if err := i.KernelDevice.DeleteAddress(addr); err != nil {
			return err
		}

		// On Darwin systems, the utun interfaces are point-to-point
		// links which are only configured a source/destination address
		// pair. Hence we need to setup dedicated routes.
		if i.KernelDevice.Flags()&net.FlagPointToPoint != 0 {
			rte := net.IPNet{
				IP:   addr.IP.Mask(addr.Mask),
				Mask: addr.Mask,
			}

			if err := i.KernelDevice.DeleteRoute(rte, i.Settings.RoutingTable); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Interface) AddAddresses(pk crypto.Key) error {
	for _, pfx := range i.Settings.Prefixes {
		addr := pk.IPAddress(pfx)
		if err := i.KernelDevice.AddAddress(addr); err != nil && !errors.Is(err, syscall.EEXIST) {
			return err
		}

		// On Darwin systems, the utun interfaces are point-to-point
		// links which are only configured a source/destination address
		// pair. Hence we need to setup dedicated routes.
		if i.KernelDevice.Flags()&net.FlagPointToPoint != 0 {
			rte := net.IPNet{
				IP:   addr.IP.Mask(addr.Mask),
				Mask: addr.Mask,
			}

			if err := i.KernelDevice.AddRoute(rte, nil, i.Settings.RoutingTable); err != nil {
				return err
			}
		}
	}

	return nil
}
