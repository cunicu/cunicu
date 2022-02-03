package intf

import (
	"context"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/pion/ice/v2"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/config"
	pice "riasc.eu/wice/internal/ice"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"
)

type Peer struct {
	wgtypes.Peer

	Interface *BaseInterface

	ConnectionState ice.ConnectionState

	agent *ice.Agent
	conn  *ice.Conn

	localOffer   *pb.Offer
	remoteOffers chan *pb.Offer

	proxy   proxy.Proxy
	backend signaling.Backend

	client             *wgctrl.Client
	config             *config.Config
	events             chan *pb.Event
	selectedCandidates chan ice.CandidatePair

	logger *zap.Logger
}

func (p *Peer) Close() error {
	if err := p.agent.Close(); err != nil {
		return fmt.Errorf("failed to close ICE agent: %w", err)
	}

	if p.proxy != nil {
		if err := p.proxy.Close(); err != nil {
			return fmt.Errorf("failed to close proxy: %w", err)
		}
	}

	p.logger.Info("Closed peer")

	return nil
}

func (p *Peer) String() string {
	return p.PublicKey().String()
}

func (p *Peer) updateEndpoint(addr *net.UDPAddr) error {
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

	if err := p.client.ConfigureDevice(p.Interface.Name(), cfg); err != nil {
		return fmt.Errorf("failed to update peer endpoint: %w", err)
	}

	p.logger.Debug("Peer endpoint updated", zap.String("endpoint", addr.String()))

	// if err := p.ensureHandshake(); err != nil {
	// 	return fmt.Errorf("failed to initiate handshake: %w", err)
	// }

	return nil
}

// isControlling determines if the peer is controlling the ICE session
// by selecting the peer which has the smaller public key
func (p *Peer) isControlling() bool {
	var pkOur, pkTheir big.Int
	pkOur.SetBytes(p.Interface.Device.PublicKey[:])
	pkTheir.SetBytes(p.Peer.PublicKey[:])

	return pkOur.Cmp(&pkTheir) == -1
}

// OnModified is a callback which gets called whenever a change of the Wireguard interface
// has been detected by the sync loop
func (p *Peer) OnModified(new *wgtypes.Peer, modified PeerModifier) {
	if modified&PeerModifiedHandshakeTime > 0 {
		p.logger.Debug("New handshake", zap.Time("time", new.LastHandshakeTime))
	}

	p.events <- &pb.Event{
		Type: pb.Event_PEER_MODIFIED,

		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &pb.Event_PeerModified{
			PeerModified: &pb.PeerModifiedEvent{
				Modified: uint32(modified),
			},
		},
	}
}

// onCandidate is a callback which gets called for each discovered local ICE candidate
func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
		return
	} else {
		p.logger.Info("Found new local candidate", zap.Any("candidate", c))

		p.localOffer.Candidates = append(p.localOffer.Candidates, pb.NewCandidate(c))
	}

	p.localOffer.Epoch++

	if err := p.backend.PublishOffer(p.PublicKeyPair(), p.localOffer); err != nil {
		p.logger.Error("Failed to publish offer", zap.Error(err))
	}
}

// onConnectionStateChange is a callback which gets called by the ICE agent
// whenever the state of the ICE connection has changed
func (p *Peer) onConnectionStateChange(state ice.ConnectionState) {
	p.ConnectionState = state

	stateLower := strings.ToLower(state.String())

	p.logger.Info("Connection state changed", zap.String("state", stateLower))

	p.events <- &pb.Event{
		Type: pb.Event_PEER_CONNECTION_STATE_CHANGED,

		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &pb.Event_PeerConnectionStateChange{
			PeerConnectionStateChange: &pb.PeerConnectionStateChangeEvent{
				NewState: pb.NewConnectionState(state),
			},
		},
	}

	if state == ice.ConnectionStateFailed {
		go func() {
			time.Sleep(p.config.RestartInterval)
			p.Restart()
		}()
	}
}

// onSelectedCandidatePairChange is a callback which gets called by the ICE agent
// whenever a new candidate pair has been selected
func (p *Peer) onSelectedCandidatePairChange(local, remote ice.Candidate) {
	p.logger.Info("Selected new candidate pair",
		zap.Any("local", local),
		zap.Any("remote", remote),
	)

	p.selectedCandidates <- ice.CandidatePair{
		Local:  local,
		Remote: remote,
	}
}

// proxyType determines a usable proxy type for the given candidate pair
func (p *Peer) proxyType(cp ice.CandidatePair) proxy.ProxyType {
	// TODO handle userspace device

	if cp.Local.Type() == ice.CandidateTypeRelay {
		return proxy.TypeUser
	} else {
		return p.config.ProxyType.ProxyType
	}
}

func (p *Peer) updateProxy(cp ice.CandidatePair) error {
	var err error

	pt := p.proxyType(cp)

	if p.proxy == nil || p.proxy.Type() != pt {
		// Close old proxy
		if p.proxy != nil {
			if err := p.proxy.Close(); err != nil {
				p.logger.Fatal("Failed to stop proxy", zap.Error(err))
			}
		}

		// Create new proxy
		if p.proxy, err = proxy.NewProxy(pt, p.ident(), p.Interface.ListenPort, p.updateEndpoint, p.conn); err != nil {
			p.logger.Fatal("Failed to setup proxy", zap.Error(err))
		}
	}

	return nil
}

// isRestartingOffer checks if a received offer should restart the
// ICE session by comparing ufrag & pwd with previously used values.
func (p *Peer) isRestartingOffer(o *pb.Offer) bool {
	ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local credentials", zap.Error(err))
	}

	credsChanged := ufrag != o.Ufrag || pwd != o.Pwd

	return p.conn != nil && credsChanged
}

// onOffer is a callback which gets called for each newly received offer
// from the remote
func (p *Peer) onOffer(o *pb.Offer) {
	p.logger.Debug("Received offer", zap.Any("offer", o))

	if p.isRestartingOffer(o) {
		p.logger.Info("Session restart triggered by remote")

		p.restart()

		if err := p.agent.SetRemoteCredentials(o.Ufrag, o.Pwd); err != nil {
			p.logger.Error("Failed to set remote creds", zap.Error(err))
			return
		}
	}

	for _, c := range o.Candidates {
		ic, err := c.ICECandidate()
		if err != nil {
			p.logger.Error("Failed to decode offer. Ignoring...", zap.Error(err))
			continue
		}

		if err := p.agent.AddRemoteCandidate(ic); err != nil {
			p.logger.Fatal("Failed to add remote candidate", zap.Error(err))
		}
		p.logger.Debug("Add remote candidate", zap.Any("candidate", c))
	}
}

// restartLocal restarts the ICE session because of a broken ICE connection
func (p *Peer) Restart() error {
	p.logger.Info("Session restart triggered locally")

	p.restart()

	return nil
}

// restart performs an ICE restart
//
// This restart can either be triggered by a failed
// ICE connection state (onConnectionState())
// or by a remote offer which indicates a restart by changed ufrag/pwd's (onOffer())
func (p *Peer) restart() {
	if err := p.agent.Restart("", ""); err != nil {
		p.logger.Error("Failed to restart ICE session", zap.Error(err))
		return
	}

	p.localOffer = p.newOffer()

	if err := p.agent.GatherCandidates(); err != nil {
		p.logger.Error("Failed to gather candidates", zap.Error(err))
		return
	}
}

func (p *Peer) start() {
	var err error

	p.logger.Info("Starting new ICE session")

	p.logger.Info("Gathering local candidates")
	if err := p.agent.GatherCandidates(); err != nil {
		p.logger.Fatal("Failed to gather candidates", zap.Error(err))
	}

	p.remoteOffers, err = p.backend.SubscribeOffers(p.PublicKeyPair())
	if err != nil {
		p.logger.Fatal("Failed to subscribe to offers", zap.Error(err))
	}

	// Wait for first offer from remote agent before creating ICE connection
	o := <-p.remoteOffers
	p.onOffer(o)

	// Start the ICE Agent. One side must be controlled, and the other must be controlling
	if p.isControlling() {
		p.conn, err = p.agent.Dial(context.Background(), o.Ufrag, o.Pwd)
	} else {
		p.conn, err = p.agent.Accept(context.Background(), o.Ufrag, o.Pwd)
	}
	if err != nil {
		p.logger.Fatal("Failed to establish ICE connection", zap.Error(err))
	}

	p.logger.Info("Connected")

	// Process more offers which are trickeling in
	for {
		select {
		case o := <-p.remoteOffers:
			if err := p.agent.SetRemoteCredentials(o.Ufrag, o.Pwd); err != nil {
				p.logger.Error("Failed to set remote creds", zap.Error(err))
				return
			}

			p.onOffer(o)

		case cp := <-p.selectedCandidates:
			if err := p.updateProxy(cp); err != nil {
				p.logger.Fatal("Failed to update proxy", zap.Error(err))
			}
		}
	}
}

// ident provides a unique string which identifies the peer
func (p *Peer) ident() string {
	return base64.StdEncoding.EncodeToString(p.Peer.PublicKey[:16])
}

// newOffer initializes a new offer before it gets send to the remote via
// the signaling backend
func (p *Peer) newOffer() *pb.Offer {
	var role pb.Offer_Role
	if p.isControlling() {
		role = pb.Offer_CONTROLLING
	} else {
		role = pb.Offer_CONTROLLED
	}

	ufrag, pwd, err := p.agent.GetLocalUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local user credentials", zap.Error(err))
	}

	return &pb.Offer{
		Version:        pb.OfferVersion,
		Type:           pb.Offer_OFFER,
		Implementation: pb.Offer_FULL,
		Role:           role,
		Epoch:          0,
		Ufrag:          ufrag,
		Pwd:            pwd,
		Candidates:     []*pb.Candidate{},
	}
}

func (p *Peer) ensureHandshake() error {
	// Return if the last handshake happed within the last 5 seconds
	if time.Since(p.LastHandshakeTime) < 5*time.Second {
		return nil
	}

	if err := p.initiateHandshake(); err != nil {
		return fmt.Errorf("failed to initiate handshake: %w", err)
	}

	return nil
}

// initiateHandshake sends a single packet towards the peer
// which triggers Wireguard to initiate the handshake
func (p *Peer) initiateHandshake() error {
	for time.Since(p.LastHandshakeTime) > 5*time.Second {
		p.logger.Debug("Waiting for handshake")

		ra := &net.UDPAddr{
			IP:   p.PublicKey().IPv6Address().IP,
			Zone: p.Interface.Name(),
			Port: 1234,
		}

		c, err := net.DialUDP("udp6", nil, ra)
		if err != nil {
			return err
		}

		if _, err := c.Write([]byte{1}); err != nil {
			return err
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (p *Peer) addAllowedIP(a net.IPNet) error {
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

// PublicKey returns the Curve25199 public key of the Wireguard peer
func (p *Peer) PublicKey() crypto.Key {
	return crypto.Key(p.Peer.PublicKey)
}

// PublicKeyPair returns both the public key of the local (our) and remote peer (theirs)
func (p *Peer) PublicKeyPair() crypto.KeyPair {
	return crypto.KeyPair{
		Ours:   p.Interface.PublicKey(),
		Theirs: p.PublicKey(),
	}
}

// PublicPrivateKeyPair returns both the public key of the local (our) and remote peer (theirs)
func (p *Peer) PublicPrivateKeyPair() crypto.KeyPair {
	return crypto.KeyPair{
		Ours:   p.Interface.PrivateKey(),
		Theirs: p.PublicKey(),
	}
}

func (p *Peer) Config() *wgtypes.PeerConfig {
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

// NewPeer creates a new peer and ICE session
func NewPeer(wgp *wgtypes.Peer, i *BaseInterface) (*Peer, error) {
	logger := zap.L().Named("peer").With(
		zap.String("intf", i.Name()),
		zap.Any("peer", wgp.PublicKey),
	)

	p := &Peer{
		Interface:          i,
		Peer:               *wgp,
		client:             i.client,
		backend:            i.backend,
		events:             i.events,
		selectedCandidates: make(chan ice.CandidatePair),
		config:             i.config,
		logger:             logger,
	}

	// Add default link-local address as allowed IP
	ap := net.IPNet{
		IP:   p.PublicKey().IPv6Address().IP,
		Mask: net.CIDRMask(128, 128),
	}
	if err := p.addAllowedIP(ap); err != nil {
		return nil, fmt.Errorf("failed to add link-local IPv6 address to AllowedIPs")
	}

	// Setup new ICE Agent
	agentConfig, err := p.config.AgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ICE agent: %w", err)
	}

	agentConfig.LoggerFactory = &pice.LoggerFactory{
		Base: p.logger,
	}

	if i.config.ProxyType.ProxyType == proxy.TypeEBPF {
		if err := proxy.SetupEBPFProxy(agentConfig, p.Interface.ListenPort); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	}

	if p.agent, err = ice.NewAgent(agentConfig); err != nil {
		return nil, fmt.Errorf("failed to create ICE agent: %w", err)
	}

	// create initial offer using freshly generated creds
	p.localOffer = p.newOffer()

	p.logger.Debug("Peer credentials",
		zap.String("ufrag", p.localOffer.Ufrag),
		zap.String("pwd", p.localOffer.Ufrag),
	)

	// When we have gathered a new ICE Candidate send it to the remote peer
	if err := p.agent.OnCandidate(p.onCandidate); err != nil {
		return nil, fmt.Errorf("failed to setup on candidate handler: %w", err)
	}

	// When selected candidate pair changes
	if err := p.agent.OnSelectedCandidatePairChange(p.onSelectedCandidatePairChange); err != nil {
		return nil, fmt.Errorf("failed to setup on selected candidate pair handler: %w", err)
	}

	// When ICE Connection state has change print to stdout
	if err := p.agent.OnConnectionStateChange(p.onConnectionStateChange); err != nil {
		return nil, fmt.Errorf("failed to setup on connection state handler: %w", err)
	}

	go p.start()

	return p, nil
}
