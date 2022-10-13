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
	daemon.Features["autocfg"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Auto-configuration",
		Order:       10,
	}
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

func (ac *Interface) Start() error {
	ac.logger.Info("Started auto-configuration")

	if err := ac.ConfigureWireGuard(); err != nil {
		ac.logger.Error("Failed to configure WireGuard interface", zap.Error(err))
	}

	// Assign auto-generated addresses
	if pk := ac.PublicKey(); pk.IsSet() {
		if err := ac.AddAddresses(pk); err != nil {
			ac.logger.Error("Failed to add addresses", zap.Error(err))
		}
	}

	// Autodetect MTU
	// TODO: Update MTU when peers are added or their endpoints change
	if mtu := ac.Settings.MTU; mtu == 0 {
		var err error
		if mtu, err = ac.DetectMTU(); err != nil {
			ac.logger.Error("Failed to detect MTU", zap.Error(err))
		} else {
			if err := ac.KernelDevice.SetMTU(mtu); err != nil {
				ac.logger.Error("Failed to set MTU", zap.Error(err), zap.Int("mtu", ac.Settings.MTU))
			}
		}
	}

	// Set link up
	if err := ac.KernelDevice.SetUp(); err != nil {
		ac.logger.Error("Failed to bring link up", zap.Error(err))
	}

	return nil
}

func (ac *Interface) Close() error {
	return nil
}

// ConfigureWireGuard configures the WireGuard device using the configuration provided by the user.
// Missing settings such as a private key or listen port are automatically generated/allocated.
func (ac *Interface) ConfigureWireGuard() error {
	var err error

	cfg := wgtypes.Config{}

	// Private key
	if !ac.PrivateKey().IsSet() || (ac.Settings.PrivateKey.IsSet() && ac.Settings.PrivateKey != ac.PrivateKey()) {
		sk := ac.Settings.PrivateKey
		if !sk.IsSet() {
			ac.logger.Warn("Device has no private key. Setting a random one.")

			sk, err = crypto.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("failed to generate private key: %w", err)
			}
		}

		cfg.PrivateKey = (*wgtypes.Key)(&sk)
	}

	// Listen port
	if ac.ListenPort == 0 || (ac.Settings.ListenPort != nil && ac.ListenPort != *ac.Settings.ListenPort) {
		if ac.ListenPort == 0 {
			ac.logger.Warn("Device has no listen port. Setting a random one.")
		}

		if ac.Settings.ListenPort != nil {
			cfg.ListenPort = ac.Settings.ListenPort
		} else {
			port, err := util.FindNextPortToListen("udp",
				ac.Settings.ListenPortRange.Min,
				ac.Settings.ListenPortRange.Max,
			)
			if err != nil {
				return fmt.Errorf("failed set listen port: %w", err)
			}

			cfg.ListenPort = &port
		}
	}

	if cfg.PrivateKey != nil || cfg.ListenPort != nil {
		if err := ac.ConfigureDevice(cfg); err != nil {
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

func (ac *Interface) RemoveAddresses(pk crypto.Key) error {
	for _, pfx := range ac.Settings.Prefixes {
		addr := pk.IPAddress(pfx)
		if err := ac.KernelDevice.DeleteAddress(addr); err != nil {
			return err
		}

		// On Darwin systems, the utun interfaces are point-to-point
		// links which are only configured a source/destination address
		// pair. Hence we need to setup dedicated routes.
		if ac.KernelDevice.Flags()&net.FlagPointToPoint != 0 {
			rte := net.IPNet{
				IP:   addr.IP.Mask(addr.Mask),
				Mask: addr.Mask,
			}

			if err := ac.KernelDevice.DeleteRoute(rte, ac.Settings.RoutingTable); err != nil {
				return err
			}
		}
	}

	return nil
}

func (ac *Interface) AddAddresses(pk crypto.Key) error {
	for _, pfx := range ac.Settings.Prefixes {
		addr := pk.IPAddress(pfx)
		if err := ac.KernelDevice.AddAddress(addr); err != nil && !errors.Is(err, syscall.EEXIST) {
			return err
		}

		if ac.KernelDevice.Flags()&net.FlagPointToPoint != 0 {
			rte := net.IPNet{
				IP:   addr.IP.Mask(addr.Mask),
				Mask: addr.Mask,
			}

			if err := ac.KernelDevice.AddRoute(rte, nil, ac.Settings.RoutingTable); err != nil {
				return err
			}
		}
	}

	return nil
}
