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

const (
	SignalingStateStable          SignalingState = iota
	SignalingStateHaveLocalOffer                 = iota
	SignalingStateHaveRemoteOffer                = iota
	SignalingStateClosed                         = iota
)

type SignalingState int

type Peer struct {
	wgtypes.Peer

	Interface *BaseInterface

	ConnectionState ice.ConnectionState

	agent *ice.Agent
	conn  *ice.Conn

	proxy   proxy.Proxy
	backend signaling.Backend

	localDescription  *pb.SessionDescription
	remoteDescription *pb.SessionDescription

	client         *wgctrl.Client
	config         *config.Config
	events         chan *pb.Event
	messages       chan *pb.SignalingMessage
	signalingState SignalingState

	logger *zap.Logger
}

// NewPeer creates a peer and initiates a new ICE agent
func NewPeer(wgp *wgtypes.Peer, i *BaseInterface) (*Peer, error) {
	logger := zap.L().Named("peer").With(
		zap.String("intf", i.Name()),
		zap.Any("peer", wgp.PublicKey),
	)

	p := &Peer{
		Interface:      i,
		Peer:           *wgp,
		client:         i.client,
		backend:        i.backend,
		events:         i.events,
		signalingState: SignalingStateStable,
		config:         i.config,
		logger:         logger,
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

	agentConfig.UDPMux = i.udpMux
	agentConfig.UDPMuxSrflx = i.udpMuxSrflx
	agentConfig.LoggerFactory = &pice.LoggerFactory{Base: p.logger}

	if p.agent, err = ice.NewAgent(agentConfig); err != nil {
		return nil, fmt.Errorf("failed to create ICE agent: %w", err)
	}

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

// Close destroys the peer as well as the ICE agent and proxies
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

// Getters

// String returns the peers public key as a base64-encoded string
func (p *Peer) String() string {
	return p.PublicKey().String()
}

// PublicKey returns the Curve25199 public key of the Wireguard peer
func (p *Peer) PublicKey() crypto.Key {
	return crypto.Key(p.Peer.PublicKey)
}

// PublicKeyPair returns both the public key of the local (our) and remote peer (theirs)
func (p *Peer) PublicKeyPair() *crypto.KeyPair {
	return &crypto.KeyPair{
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

// Config return the Wireguard peer configuration
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

// Restart performs an ICE restart
//
// This is usually triggered by a failed ICE connection state (onConnectionStateChange())
func (p *Peer) Restart() error {
	// TODO
	p.logger.Error("Restart not implemented yet!")

	return nil
}

// start starts the ICE agent
// See: https://developer.mozilla.org/en-US/docs/Web/API/WebRTC_API/Perfect_negotiation
func (p *Peer) start() {
	var err error

	// Initialize signaling channel
	if p.messages, err = p.backend.Subscribe(context.Background(), p.PublicPrivateKeyPair()); err != nil {
		p.logger.Fatal("Failed to subscribe to offers", zap.Error(err))
	}

	p.logger.Info("Starting gathering local ICE candidates")
	if err := p.agent.GatherCandidates(); err != nil {
		p.logger.Fatal("Failed to gather candidates", zap.Error(err))
	}

	if err := p.sendOffer(); err != nil {
		p.logger.Fatal("Failed to send offer", zap.Error(err))
	}

	p.signalingState = SignalingStateHaveLocalOffer

	for msg := range p.messages {
		p.onMessage(msg)
	}
}

// ident provides a unique string which identifies the peer
func (p *Peer) ident() string {
	return base64.StdEncoding.EncodeToString(p.Peer.PublicKey[:16])
}

func (p *Peer) sendOffer() error {
	p.localDescription = p.newOffer()
	msg := &pb.SignalingMessage{
		Type:        pb.SignalingMessage_OFFER,
		Description: p.localDescription,
	}

	if err := p.backend.Publish(context.Background(), p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Send offer", zap.Any("offer", p.localDescription))

	return nil
}

// newOffer initializes a new offer before it gets send via the signaling channel
func (p *Peer) newOffer() *pb.SessionDescription {
	ufrag, pwd, err := p.agent.GetLocalUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local user credentials", zap.Error(err))
	}

	return &pb.SessionDescription{
		Epoch:      1,
		Ufrag:      ufrag,
		Pwd:        pwd,
		Candidates: []*pb.Candidate{},
	}
}

// isPolite determines if the peer is controlling the ICE session
// by selecting the peer which has the smaller public key
func (p *Peer) isPolite() bool {
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

func (p *Peer) sendCandidates() {
	p.localDescription.Epoch++

	msg := &pb.SignalingMessage{
		Type:        pb.SignalingMessage_CANDIDATE,
		Description: p.localDescription,
	}

	if err := p.backend.Publish(context.Background(), p.PublicPrivateKeyPair(), msg); err != nil {
		p.logger.Error("Failed to publish offer", zap.Error(err))
	}
}

// onCandidate is a callback which gets called for each discovered local ICE candidate
func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
		return
	}

	p.logger.Info("Found new local candidate", zap.Any("candidate", c))

	p.localDescription.Candidates = append(p.localDescription.Candidates, pb.NewCandidate(c))

	if p.signalingState == SignalingStateHaveLocalOffer || p.signalingState == SignalingStateHaveRemoteOffer {
		p.sendCandidates()
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
			time.Sleep(p.config.RestartTimeout)
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
}

// proxyType determines a usable proxy type for the given candidate pair
func (p *Peer) proxyType(cp *ice.CandidatePair) proxy.ProxyType {
	// TODO handle userspace device

	if cp.Local.Type() == ice.CandidateTypeRelay {
		return proxy.TypeUser
	} else {
		return p.config.ProxyType.ProxyType
	}
}

func (p *Peer) updateProxy() error {
	var err error

	cp, err := p.agent.GetSelectedCandidatePair()
	if err != nil {
		return fmt.Errorf("failed to get selected candidate pair: %w", err)
	}

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

// updateEndpoint sets a new endpoint for the Wireguard peer
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

// addAllowedIP adds a new IP network to the allowed ip list of the Wireguard peer
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

// isSessionRestart checks if a received offer should restart the
// ICE session by comparing ufrag & pwd with previously used values.
func (p *Peer) isSessionRestart(o *pb.SessionDescription) bool {
	ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local credentials", zap.Error(err))
	}

	credsChanged := ufrag != o.Ufrag || pwd != o.Pwd

	return p.conn != nil && credsChanged
}

func (p *Peer) handleOffer(sd *pb.SessionDescription) error {
	var err error

	if err := p.agent.SetRemoteCredentials(sd.Ufrag, sd.Pwd); err != nil {
		return fmt.Errorf("failed to set remote credentials: %w", err)
	}

	answer := &pb.SignalingMessage{
		Type:        pb.SignalingMessage_ANSWER,
		Description: p.localDescription,
	}

	if err := p.backend.Publish(context.Background(), p.PublicPrivateKeyPair(), answer); err != nil {
		return fmt.Errorf("failed to publish answer: %w", err)
	}

	p.logger.Debug("Send answer", zap.Any("answer", p.localDescription))

	if p.conn, err = p.agent.Accept(context.Background(), sd.Ufrag, sd.Pwd); err != nil {
		return fmt.Errorf("failed to accept ICE connection: %w", err)
	}

	if err := p.updateProxy(); err != nil {
		return fmt.Errorf("failed to update proxy: %w", err)
	}

	return nil
}

func (p *Peer) handleAnswer(sd *pb.SessionDescription) error {
	var err error

	if err := p.agent.SetRemoteCredentials(sd.Ufrag, sd.Pwd); err != nil {
		return fmt.Errorf("failed to set remote credentials: %w", err)
	}

	if p.conn, err = p.agent.Dial(context.Background(), sd.Ufrag, sd.Pwd); err != nil {
		return fmt.Errorf("failed to dial ICE connection: %w", err)
	}

	if err := p.updateProxy(); err != nil {
		return fmt.Errorf("failed to update proxy: %w", err)
	}

	return nil
}

// handleRestart is a handler for an remotly-initiated ICE restart
func (p *Peer) handleRestart(sd *pb.SessionDescription) error {
	if err := p.agent.Restart("", ""); err != nil {
		return fmt.Errorf("failed to restart ICE session: %w", err)
	}

	p.signalingState = SignalingStateHaveRemoteOffer
	p.localDescription = p.newOffer()

	if err := p.agent.GatherCandidates(); err != nil {
		return fmt.Errorf("failed to gather candidates: %w", err)
	}

	return p.handleOffer(sd)
}

// onOffer is a handler called for each received offer via the signaling channel
func (p *Peer) onOffer(sd *pb.SessionDescription) {
	logger := p.logger.With(zap.Any("offer", sd))

	switch p.signalingState {

	case SignalingStateStable:
		logger.Info("Received offer")

		p.signalingState = SignalingStateHaveRemoteOffer

		go func() {
			if err := p.handleOffer(sd); err != nil {
				p.logger.Error("Failed to handle offer", zap.Error(err))
			}
		}()

	case SignalingStateHaveLocalOffer:
		logger.Debug("Received offer while waiting for answer")

		if p.isPolite() {
			p.signalingState = SignalingStateHaveRemoteOffer

			logger.Debug("We are polite. Accepting offer...")

			go func() {
				if err := p.handleOffer(sd); err != nil {
					logger.Error("Failed to handle offer", zap.Error(err))
				}
			}()
		} else {
			logger.Debug("We are not polite. Ignoring offer...")

			if err := p.sendOffer(); err != nil {
				logger.Error("Failed to send offer", zap.Error(err))
			}
		}

	case SignalingStateHaveRemoteOffer:
		logger.Error("Received another offer from remote")

		if p.isSessionRestart(sd) {
			logger.Info("Session restart triggered by remote")

			go func() {
				if err := p.handleRestart(sd); err != nil {
					logger.Error("Failed to handle restarting offer")
				}
			}()
		}
	}
}

// onAnswer is a handler called for each received answer via the signaling channel
func (p *Peer) onAnswer(sd *pb.SessionDescription) {
	logger := p.logger.With(zap.Any("answer", sd))

	switch p.signalingState {

	case SignalingStateStable:
		logger.Error("Received answer while not waiting for one. Ignoring...")

	case SignalingStateHaveLocalOffer:
		logger.Info("Received answer to our offer")

		p.signalingState = SignalingStateStable

		go func() {
			if err := p.handleAnswer(sd); err != nil {
				p.logger.Error("Failed to handle answer", zap.Error(err))
			}
		}()

	case SignalingStateHaveRemoteOffer:
		logger.Error("Received answer while not waiting for one. Ignoring...")
	}
}

// onMessage is called for each received message from the signaling backend
//
// Most of the negotation logic is happening here
func (p *Peer) onMessage(msg *pb.SignalingMessage) {
	switch msg.Type {

	case pb.SignalingMessage_OFFER:
		p.onOffer(msg.Description)

	case pb.SignalingMessage_ANSWER:
		p.onAnswer(msg.Description)

	case pb.SignalingMessage_CANDIDATE:

	}

	// We learn new candidates from ALL message types!
	if err := p.addCandidates(msg.Description.Candidates); err != nil {
		p.logger.Error("Failed to add candidates", zap.Error(err))
	}
}

// addCandidates deserializes a list of ICE candidates from a protobuf message
// and adds them to the ICE agent
func (p *Peer) addCandidates(cands []*pb.Candidate) error {
	for _, c := range cands {
		ic, err := c.ICECandidate()
		if err != nil {
			p.logger.Error("Failed to decode offer. Ignoring...", zap.Error(err))
			continue
		}

		if err := p.agent.AddRemoteCandidate(ic); err != nil {
			return fmt.Errorf("failed to add remote candidate: %w", err)
		}

		p.logger.Debug("Add remote candidate", zap.Any("candidate", c))
	}

	return nil
}

// ensureHandshake initiated a new Wireguard handshake if the last one is older than 5 seconds
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
