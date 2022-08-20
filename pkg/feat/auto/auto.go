package auto

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/watcher"
	"riasc.eu/wice/pkg/wg"
)

type AutoConfig struct {
	client *wgctrl.Client
	config *config.Config

	logger *zap.Logger
}

func addLinkLocalAddresses(dev device.Device, pk crypto.Key) error {
	if pk.IsSet() {
		if err := dev.AddAddress(pk.IPv4Address()); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("failed to assign IPv4 link-local address: %w", err)
		}

		if err := dev.AddAddress(pk.IPv6Address()); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("failed to assign IPv6 link-local address: %w", err)
		}
	}

	return nil
}

func deleteLinkLocalAddresses(dev device.Device, pk crypto.Key) error {
	if pk.IsSet() {
		if err := dev.DeleteAddress(pk.IPv4Address()); err != nil {
			return fmt.Errorf("failed to assign IPv4 link-local address: %w", err)
		}

		if err := dev.DeleteAddress(pk.IPv6Address()); err != nil {
			return fmt.Errorf("failed to assign IPv6 link-local address: %w", err)
		}
	}

	return nil
}

func New(w *watcher.Watcher, cfg *config.Config, client *wgctrl.Client) *AutoConfig {
	s := &AutoConfig{
		client: client,
		config: cfg,
		logger: zap.L().Named("auto"),
	}

	w.OnAll(s)

	return s
}

func (a *AutoConfig) Start() error {
	return nil
}

func (a *AutoConfig) Close() error {
	return nil
}

func (s *AutoConfig) OnInterfaceAdded(i *core.Interface) {
	logger := s.logger.With(zap.String("intf", i.Name()))

	if err := s.fixupInterface(i); err != nil {
		logger.Error("Failed to fix interface", zap.Error(err))
	}

	// Add link local addresses
	if err := addLinkLocalAddresses(i.KernelDevice, i.PublicKey()); err != nil {
		s.logger.Error("Failed to assign link-local addresses", zap.Error(err))
	}

	// Set link up
	if err := i.KernelDevice.SetUp(); err != nil {
		logger.Error("Failed to bring link up", zap.Error(err))
	}
}

func (s *AutoConfig) OnInterfaceRemoved(i *core.Interface) {}

func (s *AutoConfig) OnInterfaceModified(i *core.Interface, old *wg.Device, mod core.InterfaceModifier) {

	// Update link-local addresses in case the interface key has changed
	if mod&core.InterfaceModifiedPrivateKey != 0 {
		oldPk := crypto.Key(old.PublicKey)
		newPk := i.PublicKey()

		if err := deleteLinkLocalAddresses(i.KernelDevice, oldPk); err != nil {
			s.logger.Error("Failed to assign link-local addresses", zap.Error(err))
		}

		if err := addLinkLocalAddresses(i.KernelDevice, newPk); err != nil {
			s.logger.Error("Failed to assign link-local addresses", zap.Error(err))
		}
	}
}

func (s *AutoConfig) OnPeerAdded(p *core.Peer) {
	logger := s.logger.With(
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

func (s *AutoConfig) OnPeerRemoved(p *core.Peer) {}

func (s *AutoConfig) OnPeerModified(p *core.Peer, old *wgtypes.Peer, mod core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {

}

// fixupInterface fixes the WireGuard device configuration by applying missing settings
func (s *AutoConfig) fixupInterface(i *core.Interface) error {
	cfg := wgtypes.Config{}
	logger := s.logger.With(zap.String("intf", i.Name()))

	if !i.PrivateKey().IsSet() {
		if i.Type != wgtypes.Userspace {
			logger.Warn("Device has no private key. Generating one..")
		}

		key, err := wgtypes.GeneratePrivateKey()
		if err != nil {
			return fmt.Errorf("failed to generate private key: %w", err)
		}

		cfg.PrivateKey = &key
	}

	if i.ListenPort == 0 {
		logger.Warn("Device has no listen port. Setting a random one..")

		port, err := util.FindNextPortToListen("udp", s.config.WireGuard.Port.Min, s.config.WireGuard.Port.Max)
		if err != nil {
			return fmt.Errorf("failed set listen port: %w", err)
		}

		cfg.ListenPort = &port
	}

	if cfg.ListenPort != nil || cfg.PrivateKey != nil {
		if err := s.client.ConfigureDevice(i.Name(), cfg); err != nil {
			return fmt.Errorf("failed to configure device: %w", err)
		}
	}

	return nil
}
