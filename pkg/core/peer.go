package core

import (
	"fmt"
	"math/big"
	"net"
	"time"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/util"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
)

type SignalingState int

type Peer struct {
	*wgtypes.Peer

	Name  string
	Hosts map[string][]net.IP

	Interface *Interface

	LastReceiveTime  time.Time
	LastTransmitTime time.Time

	onModified []PeerModifiedHandler

	client *wgctrl.Client

	logger *zap.Logger
}

// NewPeer creates a peer and initiates a new ICE agent
func NewPeer(wgp *wgtypes.Peer, i *Interface) (*Peer, error) {
	p := &Peer{
		Interface: i,
		Peer:      wgp,

		onModified: []PeerModifiedHandler{},

		client: i.client,
		logger: zap.L().Named("peer").With(
			zap.String("intf", i.Name()),
			zap.Any("peer", wgp.PublicKey),
		),
	}

	// We intentionally prune the AllowedIP list here for the initial sync
	p.Peer.AllowedIPs = nil

	return p, nil
}

// Getters

// String returns the peers public key as a base64-encoded string
func (p *Peer) String() string {
	if p.Name != "" {
		return fmt.Sprintf("[%s]%s", p.Name, p.PublicKey())
	}

	return p.PublicKey().String()
}

// PublicKey returns the Curve25199 public key of the WireGuard peer
func (p *Peer) PublicKey() crypto.Key {
	return crypto.Key(p.Peer.PublicKey)
}

// PresharedKey returns the Curve25199 preshared key of the WireGuard peer
func (p *Peer) PresharedKey() crypto.Key {
	return crypto.Key(p.Peer.PresharedKey)
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

// IsControlling determines if the peer is controlling the ICE session
// by selecting the peer which has the smaller public key
func (p *Peer) IsControlling() bool {
	var pkOur, pkTheir big.Int
	pkOur.SetBytes(p.Interface.Device.PublicKey[:])
	pkTheir.SetBytes(p.Peer.PublicKey[:])

	return pkOur.Cmp(&pkTheir) == -1
}

// WireGuardConfig return the WireGuard peer configuration
func (p *Peer) WireGuardConfig() *wgtypes.PeerConfig {
	cfg := &wgtypes.PeerConfig{
		PublicKey:  *(*wgtypes.Key)(&p.Peer.PublicKey),
		Endpoint:   p.Endpoint,
		AllowedIPs: p.Peer.AllowedIPs,
	}

	if crypto.Key(p.Peer.PresharedKey).IsSet() {
		cfg.PresharedKey = &p.Peer.PresharedKey
	}

	if p.PersistentKeepaliveInterval > 0 {
		cfg.PersistentKeepaliveInterval = &p.PersistentKeepaliveInterval
	}

	return cfg
}

// OnModified registers a new handler which is called whenever the peer has been modified
func (p *Peer) OnModified(h PeerModifiedHandler) {
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

	p.logger.Debug("Peer endpoint updated", zap.Any("ep", addr))

	return nil
}

// SetPresharedKey sets a new preshared key for the WireGuard peer
func (p *Peer) SetPresharedKey(psk *crypto.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey:    p.Peer.PublicKey,
				UpdateOnly:   true,
				PresharedKey: (*wgtypes.Key)(psk),
			},
		},
	}

	if err := p.client.ConfigureDevice(p.Interface.Device.Name, cfg); err != nil {
		return fmt.Errorf("failed to update peer preshared key: %w", err)
	}

	// TODO: Remove PSK from log
	p.logger.Debug("Peer preshared key updated", zap.Any("psk", psk))

	return nil
}

// AddAllowedIP adds a new IP network to the allowed ip list of the WireGuard peer
func (p *Peer) AddAllowedIP(a net.IPNet) error {
	if util.SliceContains(p.AllowedIPs, func(n net.IPNet) bool {
		return util.CmpNet(n, a) == 0
	}) {
		p.logger.Warn("Not adding already existing allowed IP", zap.Any("ip", a))
		return nil
	}

	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				UpdateOnly: true,
				PublicKey:  wgtypes.Key(p.PublicKey()),
				AllowedIPs: []net.IPNet{a},
			},
		},
	}

	p.logger.Debug("Adding new allowed IP", zap.String("ip", a.String()))

	return p.client.ConfigureDevice(p.Interface.Device.Name, cfg)
}

// RemoveAllowedIP removes a new IP network from the allowed ip list of the WireGuard peer
func (p *Peer) RemoveAllowedIP(a net.IPNet) error {
	ips := util.SliceFilter(p.Peer.AllowedIPs, func(b net.IPNet) bool {
		return util.CmpNet(a, b) != 0
	})

	// TODO: Check is net is in AllowedIPs before attempting removing it

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

	p.logger.Debug("Remove allowed IP", zap.Any("ip", a))

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
	ipsAdded, ipsRemoved, _ := util.SliceDiffFunc(old.AllowedIPs, new.AllowedIPs, util.CmpNet)
	if len(ipsAdded) > 0 || len(ipsRemoved) > 0 {
		mod |= PeerModifiedAllowedIPs
	}

	p.Peer = new

	if mod != PeerModifiedNone {
		p.logger.Debug("Peer has been modified", zap.Strings("changes", mod.Strings()))

		for _, h := range p.onModified {
			h.OnPeerModified(p, old, mod, ipsAdded, ipsRemoved)
		}
	}

	return mod, ipsAdded, ipsRemoved
}

func (p *Peer) Marshal() *coreproto.Peer {
	allowedIPs := []string{}
	for _, allowedIP := range p.AllowedIPs {
		allowedIPs = append(allowedIPs, allowedIP.String())
	}

	q := &coreproto.Peer{
		Name:                        p.Name,
		PublicKey:                   p.PublicKey().Bytes(),
		PersistentKeepaliveInterval: uint32(p.PersistentKeepaliveInterval / time.Second),
		TransmitBytes:               p.TransmitBytes,
		ReceiveBytes:                p.ReceiveBytes,
		AllowedIps:                  allowedIPs,
		ProtocolVersion:             uint32(p.ProtocolVersion),
	}

	if p.Endpoint != nil {
		q.Endpoint = p.Endpoint.String()
	}

	if p.PresharedKey().IsSet() {
		q.PresharedKey = p.PresharedKey().Bytes()
	}

	if !p.LastHandshakeTime.IsZero() {
		q.LastHandshakeTimestamp = proto.Time(p.LastHandshakeTime)
	}

	if !p.LastReceiveTime.IsZero() {
		q.LastReceiveTimestamp = proto.Time(p.LastReceiveTime)
	}

	if !p.LastTransmitTime.IsZero() {
		q.LastTransmitTimestamp = proto.Time(p.LastTransmitTime)
	}

	return q
}
