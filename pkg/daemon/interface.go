// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"fmt"
	"io"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/log"
	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	slicesx "github.com/stv0g/cunicu/pkg/types/slices"
	"github.com/stv0g/cunicu/pkg/wg"
)

type Interface struct {
	// WireGuard handle of device
	*wg.Interface

	// OS abstractions for kernel device
	device.Device

	Peers map[crypto.Key]*Peer

	LastSync time.Time

	client *wgctrl.Client

	onModified         []InterfaceModifiedHandler
	onPeer             []PeerHandler
	onPeerStateChanged []PeerStateChangedHandler

	Daemon   *Daemon
	Settings *config.InterfaceSettings

	features map[*Feature]FeatureInterface

	logger *log.Logger
}

func NewInterface(wgDev *wgtypes.Device, client *wgctrl.Client) (*Interface, error) {
	var err error

	i := &Interface{
		Interface: (*wg.Interface)(wgDev),
		client:    client,
		Peers:     map[crypto.Key]*Peer{},

		onModified:         []InterfaceModifiedHandler{},
		onPeer:             []PeerHandler{},
		onPeerStateChanged: []PeerStateChangedHandler{},

		features: map[*Feature]FeatureInterface{},

		logger: log.Global.Named("intf").With(
			zap.String("intf", wgDev.Name),
		),
	}

	if wgDev.Type == wgtypes.Userspace {
		if i.Device, err = device.FindUserDevice(wgDev.Name); err != nil {
			return nil, fmt.Errorf("failed to find user-space WireGuard device: %w", err)
		}
	} else {
		if i.Device, err = device.FindKernelDevice(wgDev.Name); err != nil {
			return nil, fmt.Errorf("failed to find kernel-space WireGuard device: %w", err)
		}
	}

	return i, nil
}

func (i *Interface) Close() error {
	return i.ForEachFeature(func(fi FeatureInterface) error {
		return fi.Close()
	})
}

func (i *Interface) IsUserspace() bool {
	_, ok := i.Device.(*device.UserDevice)
	return ok
}

func (i *Interface) String() string {
	return i.Name()
}

func (i *Interface) Name() string {
	return i.Device.Name()
}

// PublicKey returns the Curve25519 public key of the WireGuard interface
func (i *Interface) PublicKey() crypto.Key {
	return crypto.Key(i.Interface.PublicKey)
}

// PublicKey returns the Curve25519 private key of the WireGuard interface
func (i *Interface) PrivateKey() crypto.Key {
	return crypto.Key(i.Interface.PrivateKey)
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

func (i *Interface) Marshal() *coreproto.Interface {
	return i.MarshalWithPeers(func(p *Peer) *coreproto.Peer {
		return p.Marshal()
	})
}

func (i *Interface) MarshalWithPeers(cb func(p *Peer) *coreproto.Peer) *coreproto.Interface {
	q := &coreproto.Interface{
		Name:         i.Name(),
		Type:         coreproto.InterfaceType(i.Interface.Type),
		ListenPort:   uint32(i.ListenPort),
		FirewallMark: uint32(i.FirewallMark),
		Mtu:          uint32(i.MTU()),
		Ifindex:      uint32(i.Index()),
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

func (i *Interface) Start() error {
	if err := i.ForEachFeature(func(fi FeatureInterface) error {
		return fi.Start()
	}); err != nil {
		return err
	}

	// Update Bind here so we pick up new connections and learn about the listen port
	if err := i.BindUpdate(i.ListenPort); err != nil {
		return fmt.Errorf("failed to update bind: %w", err)
	}

	return nil
}

func (i *Interface) ConfigureDevice(cfg wgtypes.Config) error {
	if err := i.Daemon.Client.ConfigureDevice(i.Name(), cfg); err != nil {
		return err
	}

	return i.Daemon.Watcher.Sync()
}

func (i *Interface) AddPeer(pcfg *wgtypes.PeerConfig) error {
	return i.ConfigureDevice(wgtypes.Config{
		Peers: []wgtypes.PeerConfig{*pcfg},
	})
}

func (i *Interface) UpdatePeer(pcfg *wgtypes.PeerConfig) error {
	pcfg2 := *pcfg
	pcfg2.UpdateOnly = true

	return i.AddPeer(&pcfg2)
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

	return i.ConfigureDevice(cfg)
}

func (i *Interface) syncInterface(newIntf *wgtypes.Device) {
	oldIntf := i.Interface
	mod := InterfaceModifiedNone

	// Compare device properties
	if newIntf.Name != oldIntf.Name {
		i.logger.Info("Name has changed",
			zap.Any("old", oldIntf.Name),
			zap.Any("new", newIntf.Name),
		)

		mod |= InterfaceModifiedName
	}

	// Compare device properties
	if newIntf.Type != oldIntf.Type {
		i.logger.Info("Type has changed",
			zap.Any("old", oldIntf.Type),
			zap.Any("new", newIntf.Type),
		)

		mod |= InterfaceModifiedType
	}

	if newIntf.FirewallMark != oldIntf.FirewallMark {
		i.logger.Info("Firewall mark has changed",
			zap.Any("old", oldIntf.FirewallMark),
			zap.Any("new", newIntf.FirewallMark),
		)

		mod |= InterfaceModifiedFirewallMark
	}

	if newIntf.PrivateKey != oldIntf.PrivateKey {
		i.logger.Info("PrivateKey has changed",
			zap.Any("old", oldIntf.PrivateKey),
			zap.Any("new", newIntf.PrivateKey),
		)

		mod |= InterfaceModifiedPrivateKey
	}

	if newIntf.ListenPort != oldIntf.ListenPort {
		i.logger.Info("ListenPort has changed",
			zap.Any("old", oldIntf.ListenPort),
			zap.Any("new", newIntf.ListenPort),
		)

		mod |= InterfaceModifiedListenPort
	}

	peersAdded, peersRemoved, peersKept := slicesx.DiffFunc(oldIntf.Peers, newIntf.Peers, wg.CmpPeers)
	if len(peersAdded) > 0 || len(peersRemoved) > 0 {
		mod |= InterfaceModifiedPeers
	}

	// Call handlers

	i.Interface = (*wg.Interface)(newIntf)
	i.LastSync = time.Now()

	if mod != InterfaceModifiedNone {
		i.logger.Debug("Interface has been modified", zap.Strings("changes", mod.Strings()))

		for _, h := range i.onModified {
			h.OnInterfaceModified(i, oldIntf, mod)
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

		// We intentionally prune the AllowedIP list here for the initial sync
		wgpCopy := *wgp
		wgp.AllowedIPs = nil

		p.Sync(&wgpCopy)
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
}

func (i *Interface) SyncFeatures() error {
	return i.ForEachFeature(func(fi FeatureInterface) error {
		if s, ok := fi.(SyncableFeatureInterface); ok {
			return s.Sync()
		}
		return nil
	})
}

func (i *Interface) BindUpdate(listenPort int) error {
	if dev, ok := i.Device.(*device.KernelDevice); ok {
		dev.ListenPort = listenPort
	}

	return i.Device.BindUpdate()
}

func (i *Interface) ForEachFeature(cb func(fi FeatureInterface) error) error {
	for _, f := range features {
		if fi, ok := i.features[f]; ok {
			if err := cb(fi); err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *Interface) OnInterfaceModified(_ *Interface, _ *wg.Interface, mod InterfaceModifier) {
	if mod&InterfaceModifiedListenPort == 0 {
		return
	}

	if err := i.BindUpdate(i.ListenPort); err != nil {
		i.logger.Error("Failed to update bind", zap.Error(err))
	}
}
