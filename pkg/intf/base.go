package intf

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/internal/wg"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"
)

type BaseInterface struct {
	wg.Device

	peers map[crypto.Key]*Peer

	lastSync time.Time

	backend signaling.Backend
	config  *config.Config
	client  *wgctrl.Client
	events  chan *pb.Event

	udpMux      ice.UDPMux
	udpMuxSrflx ice.UniversalUDPMux

	logger *zap.Logger
}

// Name returns the Wireguard interface name
func (i *BaseInterface) Name() string {
	return i.Device.Name
}

// PublicKey returns the Curve25519 public key of the Wireguard interface
func (i *BaseInterface) PublicKey() crypto.Key {
	return crypto.Key(i.Device.PublicKey)
}

// PublicKey returns the Curve25519 private key of the Wireguard interface
func (i *BaseInterface) PrivateKey() crypto.Key {
	return crypto.Key(i.Device.PrivateKey)
}

// Peers returns a map of Curve25519 public keys to Peers
func (i *BaseInterface) Peers() map[crypto.Key]*Peer {
	return i.peers
}

func (i *BaseInterface) Close() error {
	i.logger.Info("Closing interface")

	for _, p := range i.peers {
		if err := p.Close(); err != nil {
			return fmt.Errorf("failed to close peer: %w", err)
		}
	}

	return nil
}

func (i *BaseInterface) Config() *wgtypes.Config {
	cfg := &wgtypes.Config{
		PrivateKey:   (*wgtypes.Key)(i.PrivateKey().Bytes()),
		ListenPort:   &i.ListenPort,
		FirewallMark: &i.FirewallMark,
	}

	for _, peer := range i.Peers() {
		cfg.Peers = append(cfg.Peers, *peer.Config())
	}

	return cfg
}

func (i *BaseInterface) DumpConfig(wr io.Writer) error {
	cfg := wg.Config{Config: *i.Config()}
	return cfg.Dump(wr)
}

func (i *BaseInterface) Marshal() *pb.Interface {
	return pb.NewInterface((*wgtypes.Device)(&i.Device))
}

func (i *BaseInterface) syncPeer(oldPeer, newPeer *wgtypes.Peer) error {
	var modified PeerModifier = PeerModifiedNone

	// Compare peer properties
	if util.CmpEndpoint(newPeer.Endpoint, oldPeer.Endpoint) != 0 {
		modified |= PeerModifiedEndpoint
	}
	if newPeer.ProtocolVersion != oldPeer.ProtocolVersion {
		modified |= PeerModifiedProtocolVersion
	}
	if newPeer.PersistentKeepaliveInterval != oldPeer.PersistentKeepaliveInterval {
		modified |= PeerModifiedKeepaliveInterval
	}
	if newPeer.LastHandshakeTime != oldPeer.LastHandshakeTime {
		modified |= PeerModifiedHandshakeTime
	}
	if len(newPeer.AllowedIPs) != len(oldPeer.AllowedIPs) {
		modified |= PeerModifiedAllowedIPs
	} else {
		for i := 0; i < len(oldPeer.AllowedIPs); i++ {
			if util.CmpNet(&oldPeer.AllowedIPs[i], &newPeer.AllowedIPs[i]) != 0 {
				modified |= PeerModifiedAllowedIPs
				break
			}
		}
	}

	// Find changes in AllowedIP list
	// sort.Slice(newPeer.AllowedIPs, lessNets(newPeer.AllowedIPs))
	// sort.Slice(oldPeer.AllowedIPs, lessNets(oldPeer.AllowedIPs))

	// for i, j := 0, 0; i < len(oldPeer.AllowedIPs) && j < len(newPeer.AllowedIPs); {
	// 	oldAllowedIP := &oldPeer.AllowedIPs[i]
	// 	newAllowedIP := &newPeer.AllowedIPs[j]

	// 	cmp := cmpNet(oldAllowedIP, newAllowedIP)
	// 	switch {
	// 	case cmp < 0: // deleted
	// 		d.onPeerAllowedIPDeleted(oldPeer)

	// 	case cmp > 0: // added
	// 		d.onPeerAllowedIPAdded(newPeer)

	// 	default: //
	// 		i++
	// 		j++
	// 	}
	// }

	if modified != PeerModifiedNone {
		i.logger.Info("Peer modified",
			zap.Any("peer", oldPeer.PublicKey),
			zap.Any("modified", modified),
		)

		i.onPeerModified(oldPeer, newPeer, modified)
	}

	return nil
}

func (i *BaseInterface) Sync(newDev *wgtypes.Device) error {
	// Compare device properties
	if newDev.Type != i.Type {
		i.logger.Info("Type changed",
			zap.Any("old", i.Type),
			zap.Any("new", newDev.Type),
		)
		i.Device.Type = newDev.Type
	}

	if newDev.FirewallMark != i.FirewallMark {
		i.logger.Info("FirewallMark changed",
			zap.Any("old", i.FirewallMark),
			zap.Any("new", newDev.FirewallMark),
		)
		i.Device.FirewallMark = newDev.FirewallMark
	}

	if newDev.PrivateKey != i.Device.PrivateKey {
		i.logger.Info("PrivateKey changed",
			zap.Any("old", i.PrivateKey()),
			zap.Any("new", newDev.PrivateKey),
		)
		i.Device.PrivateKey = newDev.PrivateKey
		i.Device.PublicKey = newDev.PublicKey
	}

	if newDev.ListenPort != i.ListenPort {
		i.logger.Info("ListenPort changed",
			zap.Any("old", i.ListenPort),
			zap.Any("new", newDev.ListenPort),
		)
		i.Device.ListenPort = newDev.ListenPort

		// TODO: update proxy
	}

	sort.Slice(newDev.Peers, wg.LessPeers(newDev.Peers))
	sort.Slice(i.Device.Peers, wg.LessPeers(i.Device.Peers))

	k, j := 0, 0
	for k < len(i.Device.Peers) && j < len(newDev.Peers) {
		oldPeer := &i.Device.Peers[k]
		newPeer := &newDev.Peers[j]

		cmp := wg.CmpPeers(oldPeer, newPeer)
		switch {
		case cmp < 0: // removed
			i.logger.Info("Peer removed", zap.Any("peer", oldPeer.PublicKey))
			i.onPeerRemoved(oldPeer)
			k++

		case cmp > 0: // added
			i.logger.Info("Peer added", zap.Any("peer", oldPeer.PublicKey))
			i.onPeerAdded(newPeer)
			j++

		default: //
			i.syncPeer(oldPeer, newPeer)
			k++
			j++
		}
	}

	for k < len(i.Device.Peers) {
		oldPeer := &i.Device.Peers[k]
		i.logger.Info("Peer removed", zap.Any("peer", oldPeer.PublicKey))
		i.onPeerRemoved(oldPeer)
		k++
	}

	for j < len(newDev.Peers) {
		newPeer := &newDev.Peers[j]
		i.logger.Info("Peer added", zap.Any("peer", newPeer.PublicKey))
		i.onPeerAdded(newPeer)
		j++
	}

	i.Device.Peers = newDev.Peers
	i.lastSync = time.Now()

	return nil
}

func (i *BaseInterface) SyncConfig(cfgFilename string) error {
	cfgFile, err := os.Open(cfgFilename)
	if err != nil {
		return fmt.Errorf("failed to open configfile %s: %w", cfgFilename, err)
	}

	cfg, err := wg.ParseConfig(cfgFile, i.Name())
	if err != nil {
		return fmt.Errorf("failed to parse configuration: %s", err)
	}

	if err := i.client.ConfigureDevice(i.Name(), cfg.Config); err != nil {
		return fmt.Errorf("failed to sync interface config: %s", err)
	}

	// TODO: emulate wg-quick behaviour here?

	newDev, err := i.client.Device(i.Name())
	if err != nil {
		return fmt.Errorf("failed to read new config: %w", err)
	}

	i.Device = wg.Device(*newDev)

	i.logger.Debug("Synced configuration", zap.Any("config", cfg))

	return nil
}

func (i *BaseInterface) onPeerAdded(p *wgtypes.Peer) {
	peer, err := NewPeer(p, i)
	if err != nil {
		i.logger.Fatal("Failed to setup peer",
			zap.Error(err),
			zap.Any("peer", peer.PublicKey),
		)
	}

	i.peers[peer.PublicKey()] = peer

	i.events <- &pb.Event{
		Type: pb.Event_PEER_ADDED,

		Interface: i.Name(),
		Peer:      p.PublicKey[:],
	}
}

func (i *BaseInterface) onPeerRemoved(p *wgtypes.Peer) {
	peer, ok := i.peers[crypto.Key(p.PublicKey)]
	if !ok {
		i.logger.Warn("Failed to find matching peer", zap.Any("peer", peer.PublicKey))
	}

	if err := peer.Close(); err != nil {
		i.logger.Warn("Failed to close peer", zap.Any("peer", peer.PublicKey))
	}

	i.events <- &pb.Event{
		Type: pb.Event_PEER_REMOVED,

		Interface: i.Name(),
		Peer:      p.PublicKey[:],
	}

	delete(i.peers, peer.PublicKey())
}

func (i *BaseInterface) onPeerModified(old, new *wgtypes.Peer, modified PeerModifier) {
	peer, ok := i.peers[crypto.Key(new.PublicKey)]
	if ok {
		peer.OnModified(new, modified)
	} else {
		i.logger.Error("Failed to find modified peer")
	}
}

func (i *BaseInterface) AddPeer(pk wgtypes.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: pk,
			},
		},
	}

	return i.client.ConfigureDevice(i.Name(), cfg)
}

func (i *BaseInterface) RemovePeer(pk wgtypes.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: pk,
				Remove:    true,
			},
		},
	}

	return i.client.ConfigureDevice(i.Name(), cfg)
}

func (i *BaseInterface) addLinkLocalAddress() error {
	addr := i.PublicKey().IPv6Address()

	return i.addAddress(addr)
}

func NewInterface(dev *wgtypes.Device, client *wgctrl.Client, backend signaling.Backend, events chan *pb.Event, cfg *config.Config) (BaseInterface, error) {
	var err error

	logger := zap.L().Named("interface").With(
		zap.String("intf", dev.Name),
		zap.String("type", "kernel"),
	)

	i := BaseInterface{
		Device:  wg.Device(*dev),
		client:  client,
		backend: backend,
		events:  events,
		config:  cfg,
		logger:  logger,
		peers:   make(map[crypto.Key]*Peer),
	}

	i.logger.Info("Creating new interface")

	// Sync Wireguard device configuration with configuration file
	if i.config.GetBool("wg.config_sync") {
		cfg := fmt.Sprintf("%s/%s.conf", i.config.Get("wg.config_path"), i.Name())
		if err := i.SyncConfig(cfg); err != nil {
			return BaseInterface{}, fmt.Errorf("failed to sync interface configuration: %w", err)
		}
	}

	// Fixup Wireguard device configuration
	if err := i.Fixup(); err != nil {
		return BaseInterface{}, fmt.Errorf("failed to fix interface configuration: %w", err)
	}

	// Add link local address
	if err := i.addLinkLocalAddress(); err != nil {
		return BaseInterface{}, fmt.Errorf("failed to assign link-local address: %w", err)
	}

	// Create per-interface UDPMuxes
	if i.udpMux, err = proxy.CreateUDPMux(i.ListenPort); err != nil {
		return BaseInterface{}, fmt.Errorf("failed to setup UDP mux: %w", err)
	}

	if i.udpMuxSrflx, err = proxy.CreateUDPMuxSrflx(i.ListenPort + 1); err != nil {
		return BaseInterface{}, fmt.Errorf("Failed to setup UDPSrflx mux: %w", err)
	}

	i.events <- &pb.Event{
		Type: pb.Event_INTERFACE_ADDED,

		Interface: i.Name(),
	}

	// We remove all peers here so that they get added by the following sync
	i.Device.Peers = nil
	if err := i.Sync(dev); err != nil {
		return BaseInterface{}, err
	}

	return i, nil
}

func (i *BaseInterface) Fixup() error {
	var cfg wgtypes.Config

	if !i.PrivateKey().IsSet() {
		if i.Type != wgtypes.Userspace {
			i.logger.Warn("Device has no private key. Generating one..")
		}

		key, _ := wgtypes.GeneratePrivateKey()

		cfg.PrivateKey = &key
	}

	if i.ListenPort == 0 {
		i.logger.Warn("Device has no listen port. Setting a random one..")

		// Ephemeral Port Range (RFC6056 Sect. 2.1)
		portMin := (1 << 15) + (1 << 14)
		portMax := (1 << 16)
		port := portMin + rand.Intn(portMax-portMin)

		cfg.ListenPort = &port
	}

	if cfg.ListenPort != nil || cfg.PrivateKey != nil {
		if err := i.client.ConfigureDevice(i.Name(), cfg); err != nil {
			return fmt.Errorf("failed to configure device: %w", err)
		}

		// Update internal state
		if cfg.ListenPort != nil {
			i.Device.ListenPort = *cfg.ListenPort
		}

		if cfg.PrivateKey != nil {
			i.Device.PrivateKey = *cfg.PrivateKey
			i.Device.PublicKey = cfg.PrivateKey.PublicKey()
		}
	}

	return nil
}
