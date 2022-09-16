// Package autocfg handles initial auto-configuration of new interfaces and peers
package autocfg

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type AutoConfig struct {
	client *wgctrl.Client
	config *config.Config

	logger *zap.Logger
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

func New(w *watcher.Watcher, cfg *config.Config, client *wgctrl.Client) *AutoConfig {
	s := &AutoConfig{
		client: client,
		config: cfg,
		logger: zap.L().Named("autocfg"),
	}

	w.OnAll(s)

	return s
}

func (a *AutoConfig) Start() error {
	a.logger.Info("Started interface auto-configuration")

	return nil
}

func (a *AutoConfig) Close() error {
	return nil
}

func (a *AutoConfig) OnInterfaceAdded(i *core.Interface) {
	logger := a.logger.With(zap.String("intf", i.Name()))

	icfg := a.config.InterfaceSettings(i.Name())

	if err := a.configureWireGuardInterface(i, &icfg.WireGuard); err != nil {
		logger.Error("Failed to configure WireGuard interface", zap.Error(err))
	}

	// Assign addresses
	addrs := slices.Clone(icfg.AutoConfig.Addresses)
	if icfg.AutoConfig.LinkLocalAddresses && i.PublicKey().IsSet() {
		pk := i.PublicKey()
		addrs = append(addrs, pk.IPv4Address(), pk.IPv6Address())
	}

	if err := addAddresses(i.KernelDevice, addrs...); err != nil {
		logger.Error("Failed to assign addresses", zap.Error(err), zap.Any("addrs", addrs))
	}

	// Autodetect MTU
	// TODO: Update MTU when peers are added or their endpoints change
	if mtu := icfg.AutoConfig.MTU; mtu == 0 {
		var err error
		if mtu, err = i.DetectMTU(); err != nil {
			logger.Error("Failed to detect MTU", zap.Error(err))
		} else {
			if err := i.KernelDevice.SetMTU(mtu); err != nil {
				logger.Error("Failed to set MTU", zap.Error(err), zap.Int("mtu", icfg.AutoConfig.MTU))
			}
		}
	}

	// Set link up
	if err := i.KernelDevice.SetUp(); err != nil {
		logger.Error("Failed to bring link up", zap.Error(err))
	}
}

func (a *AutoConfig) OnInterfaceRemoved(i *core.Interface) {}

func (a *AutoConfig) OnInterfaceModified(i *core.Interface, old *wg.Device, mod core.InterfaceModifier) {

	// Update link-local addresses in case the interface key has changed
	if mod&core.InterfaceModifiedPrivateKey != 0 {
		oldPk := crypto.Key(old.PublicKey)
		newPk := i.PublicKey()

		if oldPk.IsSet() {
			if err := deleteAddresses(i.KernelDevice,
				oldPk.IPv4Address(),
				oldPk.IPv6Address(),
			); err != nil {
				a.logger.Error("Failed to delete link-local addresses", zap.Error(err))
			}
		}

		if newPk.IsSet() {
			if err := addAddresses(i.KernelDevice,
				newPk.IPv4Address(),
				newPk.IPv6Address(),
			); err != nil {
				a.logger.Error("Failed to assign link-local addresses", zap.Error(err))
			}
		}
	}
}

func (a *AutoConfig) OnPeerAdded(p *core.Peer) {
	logger := a.logger.With(
		zap.String("intf", p.Interface.Name()),
		zap.Any("peer", p.PublicKey()))

	// Add default link-local address as allowed IP
	ipV4 := p.PublicKey().IPv4Address()
	ipV6 := p.PublicKey().IPv6Address()

	ipV4.Mask = net.CIDRMask(32, 32)
	ipV6.Mask = net.CIDRMask(128, 128)

	if err := p.AddAllowedIP(ipV4); err != nil {
		logger.Error("Failed to add link-local IPv4 address to AllowedIPs", zap.Error(err))
	}

	if err := p.AddAllowedIP(ipV6); err != nil {
		logger.Error("Failed to add link-local IPv6 address to AllowedIPs", zap.Error(err))
	}
}

func (a *AutoConfig) OnPeerRemoved(p *core.Peer) {}

func (a *AutoConfig) OnPeerModified(p *core.Peer, old *wgtypes.Peer, mod core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
}

// configureInterface configures the WireGuard device using the configuration provided by the user
// Missing settings such as a private key or listen port are automatically generated/allocated.
func (a *AutoConfig) configureWireGuardInterface(i *core.Interface, icfg *config.WireGuardSettings) error {
	var err error

	cfg := wgtypes.Config{}
	configure := false

	logger := a.logger.With(zap.String("intf", i.Name()))

	// Private key
	if !i.PrivateKey().IsSet() || i.PrivateKey() != icfg.PrivateKey {
		sk := icfg.PrivateKey
		if !sk.IsSet() {
			sk, err = crypto.GeneratePrivateKey()
			if err != nil {
				return fmt.Errorf("failed to generate private key: %w", err)
			}
		}

		cfg.PrivateKey = (*wgtypes.Key)(&sk)
		configure = true
	}

	// Listen port
	if i.ListenPort == 0 || i.ListenPort != *icfg.ListenPort {
		if icfg.ListenPort != nil {
			cfg.ListenPort = icfg.ListenPort
		} else {
			logger.Warn("Device has no listen port. Setting a random one.")

			port, err := util.FindNextPortToListen("udp",
				icfg.ListenPortRange.Min,
				icfg.ListenPortRange.Max,
			)
			if err != nil {
				return fmt.Errorf("failed set listen port: %w", err)
			}

			cfg.ListenPort = &port
		}

		configure = true
	}

	if configure {
		if err := a.client.ConfigureDevice(i.Name(), cfg); err != nil {
			return fmt.Errorf("failed to configure device: %w", err)
		}
	}

	return nil
}
