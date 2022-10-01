// Package autocfg handles initial auto-configuration of new interfaces and peers
package autocfg

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/util"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func init() {
	daemon.Features["autocfg"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Interface auto-configuration",
		Order:       10,
	}
}

type Interface struct {
	*daemon.Interface

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.AutoConfig.Enabled {
		return nil, nil
	}

	a := &Interface{
		Interface: i,
		logger:    zap.L().Named("autocfg").With(zap.String("intf", i.Name())),
	}

	i.OnModified(a)
	i.OnPeer(a)

	return a, nil
}

func (a *Interface) Start() error {
	a.logger.Info("Started interface auto-configuration")

	if err := a.configureWireGuardInterface(); err != nil {
		a.logger.Error("Failed to configure WireGuard interface", zap.Error(err))
	}

	// Assign addresses
	addrs := slices.Clone(a.Settings.AutoConfig.Addresses)
	if a.Settings.AutoConfig.LinkLocalAddresses && a.PublicKey().IsSet() {
		addrs = append(addrs, a.PublicKey().IPv4Address(), a.PublicKey().IPv6Address())
	}

	if err := addAddresses(a.KernelDevice, addrs...); err != nil {
		a.logger.Error("Failed to assign addresses", zap.Error(err), zap.Any("addrs", addrs))
	}

	// Autodetect MTU
	// TODO: Update MTU when peers are added or their endpoints change
	if mtu := a.Settings.AutoConfig.MTU; mtu == 0 {
		var err error
		if mtu, err = a.DetectMTU(); err != nil {
			a.logger.Error("Failed to detect MTU", zap.Error(err))
		} else {
			if err := a.KernelDevice.SetMTU(mtu); err != nil {
				a.logger.Error("Failed to set MTU", zap.Error(err), zap.Int("mtu", a.Settings.AutoConfig.MTU))
			}
		}
	}

	// Set link up
	if err := a.KernelDevice.SetUp(); err != nil {
		a.logger.Error("Failed to bring link up", zap.Error(err))
	}

	return nil
}

func (a *Interface) Close() error {
	return nil
}

// configureWireGuardInterface configures the WireGuard device using the configuration provided by the user.
// Missing settings such as a private key or listen port are automatically generated/allocated.
func (a *Interface) configureWireGuardInterface() error {
	var err error

	cfg := wgtypes.Config{}

	// Private key
	if !a.PrivateKey().IsSet() || (a.Settings.WireGuard.PrivateKey.IsSet() && a.Settings.WireGuard.PrivateKey != a.PrivateKey()) {
		sk := a.Settings.WireGuard.PrivateKey
		if !sk.IsSet() {
			a.logger.Warn("Device has no private key. Setting a random one.")

			sk, err = crypto.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("failed to generate private key: %w", err)
			}
		}

		cfg.PrivateKey = (*wgtypes.Key)(&sk)
	}

	// Listen port
	if a.ListenPort == 0 || (a.Settings.WireGuard.ListenPort != nil && a.ListenPort != *a.Settings.WireGuard.ListenPort) {
		if a.ListenPort == 0 {
			a.logger.Warn("Device has no listen port. Setting a random one.")
		}

		if a.Settings.WireGuard.ListenPort != nil {
			cfg.ListenPort = a.Settings.WireGuard.ListenPort
		} else {
			port, err := util.FindNextPortToListen("udp",
				a.Settings.WireGuard.ListenPortRange.Min,
				a.Settings.WireGuard.ListenPortRange.Max,
			)
			if err != nil {
				return fmt.Errorf("failed set listen port: %w", err)
			}

			cfg.ListenPort = &port
		}
	}

	if cfg.PrivateKey != nil || cfg.ListenPort != nil {
		if err := a.Daemon.ConfigureDevice(a.Name(), cfg); err != nil {
			return fmt.Errorf("failed to configure device: %w", err)
		}
	}

	return nil
}

func addAddresses(dev device.Device, addrs ...net.IPNet) error {
	for _, addr := range addrs {
		if err := dev.AddAddress(addr); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("failed to assign address: %w", err)
		}
	}

	return nil
}

func deleteAddresses(dev device.Device, addrs ...net.IPNet) error {
	for _, addr := range addrs {
		if err := dev.DeleteAddress(addr); err != nil {
			return fmt.Errorf("failed to assign IPv4 link-local address: %w", err)
		}
	}

	return nil
}
