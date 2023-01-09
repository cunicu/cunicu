package core

import (
	"fmt"
	"io"
	"time"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/device"
	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Interface struct {
	// WireGuard handle of device
	*wg.Device

	// OS abstractions for kernel device
	KernelDevice device.Device

	Peers map[crypto.Key]*Peer

	LastSync time.Time

	client *wgctrl.Client

	onModified []InterfaceModifiedHandler
	onPeer     []PeerHandler

	logger *zap.Logger
}

func NewInterface(wgDev *wgtypes.Device, client *wgctrl.Client) (*Interface, error) {
	var err error

	i := &Interface{
		Device: (*wg.Device)(wgDev),
		client: client,
		Peers:  map[crypto.Key]*Peer{},

		onModified: []InterfaceModifiedHandler{},
		onPeer:     []PeerHandler{},

		logger: zap.L().Named("intf").With(
			zap.String("intf", wgDev.Name),
		),
	}

	if i.KernelDevice, err = device.FindDevice(wgDev.Name); err != nil {
		return nil, fmt.Errorf("failed to find kernel device: %w", err)
	}

	i.logger.Info("Added interface",
		zap.Any("pk", i.PublicKey()),
		zap.Any("type", i.Type),
		zap.Int("num_peers", len(i.Peers)),
	)

	return i, nil
}

func (i *Interface) String() string {
	return i.Device.Name
}

func (i *Interface) OnModified(h InterfaceModifiedHandler) {
	i.onModified = append(i.onModified, h)
}

func (i *Interface) OnPeer(h PeerHandler) {
	i.onPeer = append(i.onPeer, h)
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

func (i *Interface) Sync(newDev *wgtypes.Device) (InterfaceModifier, []wgtypes.Peer, []wgtypes.Peer) {
	oldDev := i.Device
	mod := InterfaceModifiedNone

	// Compare device properties
	if newDev.Name != oldDev.Name {
		i.logger.Info("Name has changed",
			zap.Any("old", oldDev.Name),
			zap.Any("new", newDev.Name),
		)

		mod |= InterfaceModifiedName
	}

	// Compare device properties
	if newDev.Type != oldDev.Type {
		i.logger.Info("Type has changed",
			zap.Any("old", oldDev.Type),
			zap.Any("new", newDev.Type),
		)

		mod |= InterfaceModifiedType
	}

	if newDev.FirewallMark != oldDev.FirewallMark {
		i.logger.Info("Firewall mark has changed",
			zap.Any("old", oldDev.FirewallMark),
			zap.Any("new", newDev.FirewallMark),
		)

		mod |= InterfaceModifiedFirewallMark
	}

	if newDev.PrivateKey != oldDev.PrivateKey {
		i.logger.Info("PrivateKey has changed",
			zap.Any("old", oldDev.PrivateKey),
			zap.Any("new", newDev.PrivateKey),
		)

		mod |= InterfaceModifiedPrivateKey
	}

	if newDev.ListenPort != oldDev.ListenPort {
		i.logger.Info("ListenPort has changed",
			zap.Any("old", oldDev.ListenPort),
			zap.Any("new", newDev.ListenPort),
		)

		mod |= InterfaceModifiedListenPort
	}

	peersAdded, peersRemoved, peersKept := util.SliceDiffFunc(oldDev.Peers, newDev.Peers, wg.CmpPeers)
	if len(peersAdded) > 0 || len(peersRemoved) > 0 {
		mod |= InterfaceModifiedPeers
	}

	// Call handlers

	i.Device = (*wg.Device)(newDev)
	i.LastSync = time.Now()

	if mod != InterfaceModifiedNone {
		i.logger.Debug("Interface has been modified", zap.Strings("changes", mod.Strings()))

		for _, h := range i.onModified {
			h.OnInterfaceModified(i, oldDev, mod)
		}
	}

	for j := range peersRemoved {
		wgp := peersRemoved[j]
		pk := crypto.Key(wgp.PublicKey)

		p, ok := i.Peers[pk]
		if !ok {
			i.logger.Warn("Failed to find matching peer", zap.Any("peer", wgp.PublicKey))
			continue
		}

		i.logger.Info("Removed peer", zap.Any("peer", p.PublicKey()))

		delete(i.Peers, pk)

		for _, h := range i.onPeer {
			h.OnPeerRemoved(p)
		}
	}

	for j := range peersAdded {
		wgp := &peersAdded[j]

		i.logger.Info("Added peer", zap.Any("peer", wgp.PublicKey))

		p, err := NewPeer(wgp, i)
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

		p.Sync(wgp)
	}

	for j := range peersKept {
		wgp := &peersKept[j]

		p, ok := i.Peers[crypto.Key(wgp.PublicKey)]
		if !ok {
			i.logger.Warn("Failed to find matching peer", zap.Any("peer", wgp.PublicKey))
			continue
		}

		p.Sync(wgp)
	}

	return mod, peersAdded, peersRemoved
}

func (i *Interface) Marshal() *coreproto.Interface {
	return i.MarshalWithPeers(func(p *Peer) *coreproto.Peer {
		return p.Marshal()
	})
}

func (i *Interface) MarshalWithPeers(cb func(p *Peer) *coreproto.Peer) *coreproto.Interface {
	q := &coreproto.Interface{
		Name:         i.Name(),
		Type:         coreproto.InterfaceType(i.Type),
		ListenPort:   uint32(i.ListenPort),
		FirewallMark: uint32(i.FirewallMark),
		Mtu:          uint32(i.KernelDevice.MTU()),
		Ifindex:      uint32(i.KernelDevice.Index()),
	}

	if cb == nil {
		cb = func(p *Peer) *coreproto.Peer {
			return p.Marshal()
		}
	}

	for _, p := range i.Peers {
		if qp := cb(p); qp != nil {
			q.Peers = append(q.Peers, qp)
		}
	}

	if !i.LastSync.IsZero() {
		q.LastSyncTimestamp = proto.Time(i.LastSync)
	}

	if i.PrivateKey().IsSet() {
		q.PrivateKey = i.PrivateKey().Bytes()
		q.PublicKey = i.PublicKey().Bytes()
	}

	return q
}
