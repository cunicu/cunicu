package core

import (
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/wg"
)

type Interface struct {
	// WireGuard handle of device
	wg.Device

	// OS abstractions for kernel device
	KernelDevice device.KernelDevice

	Peers map[crypto.Key]*Peer

	LastSync time.Time

	client *wgctrl.Client

	onModified []InterfaceHandler
	onPeer     []PeerHandler

	logger *zap.Logger
}

func (i *Interface) OnModified(h InterfaceHandler) {
	i.onModified = append(i.onModified, h)
}

func (i *Interface) OnPeer(h PeerHandler) {
	i.onPeer = append(i.onPeer, h)
}

func (i *Interface) Close() error {
	i.logger.Info("Closing interface")

	if err := i.KernelDevice.Close(); err != nil {
		return err
	}

	return nil
}

// Name returns the WireGuard interface name
func (i *Interface) Name() string {
	return i.Device.Name
}

// PublicKey returns the Curve25519 public key of the WireGuard interface
func (i *Interface) PublicKey() crypto.Key {
	return crypto.Key(i.Device.PublicKey)
}

// PublicKey returns the Curve25519 private key of the WireGuard interface
func (i *Interface) PrivateKey() crypto.Key {
	return crypto.Key(i.Device.PrivateKey)
}

func (i *Interface) WireGuardConfig() *wgtypes.Config {
	cfg := &wgtypes.Config{
		PrivateKey:   (*wgtypes.Key)(i.PrivateKey().Bytes()),
		ListenPort:   &i.ListenPort,
		FirewallMark: &i.FirewallMark,
	}

	for _, peer := range i.Peers {
		cfg.Peers = append(cfg.Peers, *peer.WireGuardConfig())
	}

	return cfg
}

func (i *Interface) DumpConfig(wr io.Writer) error {
	cfg := wg.Config{
		Config: *i.WireGuardConfig(),
	}
	return cfg.Dump(wr)
}

func (i *Interface) Marshal() *pb.Interface {
	return pb.NewInterface((*wgtypes.Device)(&i.Device))
}

func (i *Interface) Sync(new *wgtypes.Device) (InterfaceModifier, []wgtypes.Peer, []wgtypes.Peer) {
	old := i.Device
	mod := InterfaceModifiedNone

	// Compare device properties
	if new.Name != old.Name {
		i.logger.Info("Name changed",
			zap.Any("old", old.Name),
			zap.Any("new", new.Name),
		)

		mod |= InterfaceModifiedName
	}

	// Compare device properties
	if new.Type != old.Type {
		i.logger.Info("Type changed",
			zap.Any("old", old.Type),
			zap.Any("new", new.Type),
		)

		mod |= InterfaceModifiedType
	}

	if new.FirewallMark != old.FirewallMark {
		i.logger.Info("FirewallMark changed",
			zap.Any("old", old.FirewallMark),
			zap.Any("new", new.FirewallMark),
		)

		mod |= InterfaceModifiedFirewallMark
	}

	if new.PrivateKey != old.PrivateKey {
		i.logger.Info("PrivateKey changed",
			zap.Any("old", old.PrivateKey),
			zap.Any("new", new.PrivateKey),
		)

		mod |= InterfaceModifiedPrivateKey
	}

	if new.ListenPort != old.ListenPort {
		i.logger.Info("ListenPort changed",
			zap.Any("old", old.ListenPort),
			zap.Any("new", new.ListenPort),
		)

		mod |= InterfaceModifiedListenPort
	}

	peersAdded, peersRemoved, peersKept := util.DiffSliceFunc(old.Peers, new.Peers, wg.CmpPeers)
	if len(peersAdded) > 0 || len(peersRemoved) > 0 {
		mod |= InterfaceModifiedPeers
	}

	// Call handlers

	i.Device = wg.Device(*new)
	i.LastSync = time.Now()

	if mod != InterfaceModifiedNone {
		i.logger.Info("Interface modified", zap.Strings("modified", mod.Strings()))

		for _, h := range i.onModified {
			h.OnInterfaceModified(i, &old, mod)
		}
	}

	for _, wgp := range peersRemoved {
		p, ok := i.Peers[crypto.Key(wgp.PublicKey)]
		if !ok {
			i.logger.Warn("Failed to find matching peer", zap.Any("peer", wgp.PublicKey))
			continue
		}

		i.logger.Info("Peer removed", zap.Any("peer", p.PublicKey()))

		delete(i.Peers, p.PublicKey())

		for _, h := range i.onPeer {
			h.OnPeerRemoved(p)
		}
	}

	for _, wgp := range peersAdded {
		i.logger.Info("Peer added", zap.Any("peer", wgp.PublicKey))

		p, err := NewPeer(&wgp, i)
		if err != nil {
			i.logger.Fatal("Failed to setup peer",
				zap.Error(err),
				zap.Any("peer", p.PublicKey),
			)
		}

		i.Peers[p.PublicKey()] = p

		for _, h := range i.onPeer {
			h.OnPeerAdded(p)
		}

		p.Sync(&wgp)
	}

	for _, wgp := range peersKept {
		p, ok := i.Peers[crypto.Key(wgp.PublicKey)]
		if !ok {
			i.logger.Warn("Failed to find matching peer", zap.Any("peer", wgp.PublicKey))
			continue
		}

		p.Sync(&wgp)
	}

	return mod, peersAdded, peersRemoved
}

func (i *Interface) SyncConfig(cfgFilename string) error {
	cfgFile, err := os.Open(cfgFilename)
	if err != nil {
		return fmt.Errorf("failed to open config file %s: %w", cfgFilename, err)
	}

	cfg, err := wg.ParseConfig(cfgFile, i.Name())
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %s", err)
	}

	if err := i.Configure(cfg.Config); err != nil {
		return fmt.Errorf("failed to configure interface: %w", err)
	}

	i.logger.Info("Synchronized configuration", zap.String("config_file", cfgFilename))

	return nil
}

func (i *Interface) Configure(cfg wgtypes.Config) error {
	if err := i.client.ConfigureDevice(i.Name(), cfg); err != nil {
		return fmt.Errorf("failed to sync interface config: %s", err)
	}

	// TODO: Emulate wg-quick behavior here?
	//       E.g. all Pre/Post-Up/Down scripts

	return nil
}

func (i *Interface) AddPeer(pk crypto.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: wgtypes.Key(pk),
			},
		},
	}

	return i.client.ConfigureDevice(i.Name(), cfg)
}

func (i *Interface) RemovePeer(pk crypto.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: wgtypes.Key(pk),
				Remove:    true,
			},
		},
	}

	return i.client.ConfigureDevice(i.Name(), cfg)
}

func NewInterface(wgDev *wgtypes.Device, kDev device.KernelDevice, client *wgctrl.Client) (*Interface, error) {
	var err error

	logger := zap.L().Named("interface").With(
		zap.String("intf", wgDev.Name),
	)

	if kDev == nil {
		kDev, err = device.FindDevice(wgDev.Name)
		if err != nil {
			return nil, err
		}
	}

	i := &Interface{
		Device:       wg.Device(*wgDev),
		KernelDevice: kDev,
		client:       client,
		logger:       logger,
		Peers:        map[crypto.Key]*Peer{},

		onModified: []InterfaceHandler{},
		onPeer:     []PeerHandler{},
	}

	// We purposefully prune the peer list here for an full initial sync of all peers
	i.Device.Peers = nil

	i.logger.Info("Added new interface")

	return i, nil
}

func CreateInterface(name string, user bool, client *wgctrl.Client) (*Interface, error) {
	var newDevice func(name string) (device.KernelDevice, error)
	if user {
		newDevice = device.NewUserDevice
	} else {
		newDevice = device.NewKernelDevice
	}

	kDev, err := newDevice(name)
	if err != nil {
		return nil, err
	}

	// Connect to UAPI
	wgDev, err := client.Device(name)
	if err != nil {
		return nil, err
	}

	i, err := NewInterface(wgDev, kDev, client)
	if err != nil {
		return nil, err
	}

	return i, nil
}
