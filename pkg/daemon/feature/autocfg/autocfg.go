// Package autocfg handles initial auto-configuration of new interfaces and peers
package autocfg

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/util"
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

func (ac *Interface) Start() error {
	ac.logger.Info("Started auto-configuration")

	if err := ac.ConfigureWireGuard(); err != nil {
		ac.logger.Error("Failed to configure WireGuard interface", zap.Error(err))
	}

	// Assign auto-generated addresses
	if sk := ac.PrivateKey(); sk.IsSet() {
		pk := sk.PublicKey()
		if err := ac.AddAddresses(pk); err != nil {
			ac.logger.Error("Failed to add addresses", zap.Error(err))
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

		// On Darwin systems, the utun interfaces are point-to-point
		// links which are only configured a source/destination address
		// pair. Hence we need to setup dedicated routes.
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
