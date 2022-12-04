package core

import (
	"fmt"
	"math/big"
	"net"
	"time"

	"github.com/stv0g/cunicu/pkg/crypto"
	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	"github.com/stv0g/cunicu/pkg/util"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
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
	}

	p.logger = zap.L().Named("peer").With(
		zap.String("intf", i.Name()),
		zap.String("peer", p.String()),
	)

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
		PublicKey:  p.Peer.PublicKey,
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

// SetEndpoint sets a new endpoint for the WireGuard peer
func (p *Peer) SetEndpoint(addr *net.UDPAddr) error {
	// Check if update is required
	if util.CmpUDPAddr(addr, p.Endpoint) == 0 {
		return nil
	}

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

	p.logger.Debug("Peer endpoint changed", zap.Any("ep", addr))

	return nil
}

// SetPresharedKey sets a new preshared key for the WireGuard peer
func (p *Peer) SetPresharedKey(psk *crypto.Key) error {
	// Check if update is required
	if *psk == p.PresharedKey() {
		return nil
	}

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

	p.logger.Debug("Peer preshared key updated")

	return nil
}

// AddAllowedIP adds a new IP network to the allowed ip list of the WireGuard peer
func (p *Peer) AddAllowedIP(a net.IPNet) error {
	// Check if AllowedIP is already configured
	if util.SliceContains(p.AllowedIPs, func(n net.IPNet) bool {
		return util.CmpNet(n, a) == 0
	}) {
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

	// Check if AllowedIP is configured
	if len(ips) == len(p.Peer.AllowedIPs) {
		return nil
	}

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

func (p *Peer) Sync(newPeer *wgtypes.Peer) (PeerModifier, []net.IPNet, []net.IPNet) {
	oldPeer := p.Peer
	mod := PeerModifiedNone

	now := time.Now()

	// Compare peer properties
	if newPeer.PresharedKey != oldPeer.PresharedKey {
		mod |= PeerModifiedPresharedKey
	}
	if util.CmpUDPAddr(newPeer.Endpoint, oldPeer.Endpoint) != 0 {
		mod |= PeerModifiedEndpoint
	}
	if newPeer.PersistentKeepaliveInterval != oldPeer.PersistentKeepaliveInterval {
		mod |= PeerModifiedKeepaliveInterval
	}
	if newPeer.LastHandshakeTime != oldPeer.LastHandshakeTime {
		mod |= PeerModifiedHandshakeTime
	}
	if newPeer.ReceiveBytes != oldPeer.ReceiveBytes {
		mod |= PeerModifiedReceiveBytes
		p.LastReceiveTime = now
	}
	if newPeer.TransmitBytes != oldPeer.TransmitBytes {
		mod |= PeerModifiedTransmitBytes
		p.LastTransmitTime = now
	}
	if newPeer.ProtocolVersion != oldPeer.ProtocolVersion {
		mod |= PeerModifiedProtocolVersion
	}

	// Find changes in AllowedIP list
	ipsAdded, ipsRemoved, _ := util.SliceDiffFunc(oldPeer.AllowedIPs, newPeer.AllowedIPs, util.CmpNet)
	if len(ipsAdded) > 0 || len(ipsRemoved) > 0 {
		mod |= PeerModifiedAllowedIPs
	}

	p.Peer = newPeer

	if mod != PeerModifiedNone {
		p.logger.Debug("Peer has been modified", zap.Strings("changes", mod.Strings()))

		for _, h := range p.onModified {
			h.OnPeerModified(p, oldPeer, mod, ipsAdded, ipsRemoved)
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
		Reachability:                p.Reachability(),
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

func (p *Peer) Reachability() coreproto.ReachabilityType {
	if p.Endpoint == nil {
		return coreproto.ReachabilityType_REACHABILITY_TYPE_NONE
	}

	now := time.Now()
	lastActivity := p.LastReceiveTime
	if p.LastTransmitTime.After(lastActivity) {
		lastActivity = p.LastTransmitTime
	}

	//nolint:gocritic
	if p.LastHandshakeTime.After(now.Add(-2 * time.Minute)) {
		return coreproto.ReachabilityType_REACHABILITY_TYPE_DIRECT
	} else if lastActivity.After(p.LastHandshakeTime) {
		return coreproto.ReachabilityType_REACHABILITY_TYPE_NONE
	} else {
		return coreproto.ReachabilityType_REACHABILITY_TYPE_UNKNOWN
	}
}
