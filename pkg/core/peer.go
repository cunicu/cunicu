package core

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/util"
)

type SignalingState int

type Peer struct {
	wgtypes.Peer

	Name string

	Interface *Interface

	LastReceiveTime  time.Time
	LastTransmitTime time.Time

	onModified []PeerHandler

	client *wgctrl.Client

	logger *zap.Logger
}

// NewPeer creates a peer and initiates a new ICE agent
func NewPeer(wgp *wgtypes.Peer, i *Interface) (*Peer, error) {
	logger := zap.L().Named("peer").With(
		zap.String("intf", i.Name()),
		zap.Any("peer", wgp.PublicKey),
	)

	p := &Peer{
		Interface: i,
		Peer:      *wgp,

		onModified: []PeerHandler{},

		client: i.client,
		logger: logger,
	}

	// We intentionally prune the AllowedIP list here for the initial sync
	p.Peer.AllowedIPs = nil

	return p, nil
}

// Getters

// String returns the peers public key as a base64-encoded string
func (p *Peer) String() string {
	if p.Name != "" {
		return fmt.Sprintf("%s[%s]", p.Name, p.PublicKey().String())
	} else {
		return fmt.Sprintf("[%s]", p.PublicKey().String())
	}
}

// PublicKey returns the Curve25199 public key of the WireGuard peer
func (p *Peer) PublicKey() crypto.Key {
	return crypto.Key(p.Peer.PublicKey)
}

// PublicKeyPair returns both the public key of the local (our) and remote peer (theirs)
func (p *Peer) PublicKeyPair() *crypto.PublicKeyPair {
	return &crypto.PublicKeyPair{
		Ours:   p.Interface.PublicKey(),
		Theirs: p.PublicKey(),
	}
}

// PublicPrivateKeyPair returns both the public key of the local (our) and remote peer (theirs)
func (p *Peer) PublicPrivateKeyPair() *crypto.KeyPair {
	return &crypto.KeyPair{
		Ours:   p.Interface.PrivateKey(),
		Theirs: p.PublicKey(),
	}
}

// PeerConfig return the WireGuard peer configuration
func (p *Peer) WireGuardConfig() *wgtypes.PeerConfig {
	cfg := &wgtypes.PeerConfig{
		PublicKey:  *(*wgtypes.Key)(&p.Peer.PublicKey),
		Endpoint:   p.Endpoint,
		AllowedIPs: p.Peer.AllowedIPs,
	}

	if crypto.Key(p.PresharedKey).IsSet() {
		cfg.PresharedKey = &p.PresharedKey
	}

	if p.PersistentKeepaliveInterval > 0 {
		cfg.PersistentKeepaliveInterval = &p.PersistentKeepaliveInterval
	}

	return cfg
}

func (p *Peer) OnModified(h PeerHandler) {
	p.onModified = append(p.onModified, h)
}

// UpdateEndpoint sets a new endpoint for the WireGuard peer
func (p *Peer) UpdateEndpoint(addr *net.UDPAddr) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:         p.Peer.PublicKey,
				UpdateOnly:        true,
				ReplaceAllowedIPs: false,
				Endpoint:          addr,
			},
		},
	}

	if err := p.client.ConfigureDevice(p.Interface.Device.Name, cfg); err != nil {
		return fmt.Errorf("failed to update peer endpoint: %w", err)
	}

	p.logger.Debug("Peer endpoint updated", zap.Any("endpoint", addr))

	return nil
}

// AddAllowedIP adds a new IP network to the allowed ip list of the WireGuard peer
func (p *Peer) AddAllowedIP(a *net.IPNet) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				UpdateOnly: true,
				PublicKey:  wgtypes.Key(p.PublicKey()),
				AllowedIPs: []net.IPNet{*a},
			},
		},
	}

	p.logger.Debug("Adding new allowed IP", zap.String("ip", a.String()))

	return p.client.ConfigureDevice(p.Interface.Device.Name, cfg)
}

// RemoveAllowedIP removes a new IP network from the allowed ip list of the WireGuard peer
func (p *Peer) RemoveAllowedIP(a *net.IPNet) error {
	ips := util.FilterSlice(p.Peer.AllowedIPs, func(b net.IPNet) bool {
		return util.CmpNet(a, &b) != 0
	})

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				UpdateOnly:        true,
				PublicKey:         wgtypes.Key(p.PublicKey()),
				ReplaceAllowedIPs: true,
				AllowedIPs:        ips,
			},
		},
	}

	p.logger.Debug("Adding new allowed IP", zap.String("ip", a.String()))

	return p.client.ConfigureDevice(p.Interface.Device.Name, cfg)
}

func (p *Peer) Sync(new *wgtypes.Peer) (PeerModifier, []net.IPNet, []net.IPNet) {
	old := p.Peer
	mod := PeerModifiedNone

	now := time.Now()

	// Compare peer properties
	if new.PresharedKey != old.PresharedKey {
		mod |= PeerModifiedPresharedKey
	}
	if util.CmpEndpoint(new.Endpoint, old.Endpoint) != 0 {
		mod |= PeerModifiedEndpoint
	}
	if new.PersistentKeepaliveInterval != old.PersistentKeepaliveInterval {
		mod |= PeerModifiedKeepaliveInterval
	}
	if new.LastHandshakeTime != old.LastHandshakeTime {
		mod |= PeerModifiedHandshakeTime
	}
	if new.ReceiveBytes != old.ReceiveBytes {
		mod |= PeerModifiedReceiveBytes
		p.LastReceiveTime = now
	}
	if new.TransmitBytes != old.TransmitBytes {
		mod |= PeerModifiedTransmitBytes
		p.LastTransmitTime = now
	}
	if new.ProtocolVersion != old.ProtocolVersion {
		mod |= PeerModifiedProtocolVersion
	}

	// Find changes in AllowedIP list
	ipsAdded, ipsRemoved, _ := util.DiffSliceFunc(old.AllowedIPs, new.AllowedIPs, util.CmpNet)
	if len(ipsAdded) > 0 || len(ipsRemoved) > 0 {
		mod |= PeerModifiedAllowedIPs
	}

	p.Peer = *new

	if mod != PeerModifiedNone {
		p.logger.Info("Peer modified", zap.Strings("modified", mod.Strings()))

		for _, h := range p.onModified {
			h.OnPeerModified(p, &old, mod, ipsAdded, ipsRemoved)
		}
	}

	return mod, ipsAdded, ipsRemoved
}
