package intf

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"os/exec"
	"strings"

	"go.uber.org/zap"

	"github.com/pion/ice/v2"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/internal/log"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"
)

const (
	ConnectionStateConnecting = 100
)

type SignalingState int

type Peer struct {
	wgtypes.Peer

	Interface *BaseInterface

	ConnectionState ice.ConnectionState

	agent *ice.Agent
	conn  *ice.Conn

	proxy   *proxy.Proxy
	backend signaling.Backend

	description *pb.SessionDescription

	client    *wgctrl.Client
	config    *config.Config
	events    chan *pb.Event
	messages  chan *pb.SignalingMessage
	doRestart chan interface{}

	logger *zap.Logger
}

// NewPeer creates a peer and initiates a new ICE agent
func NewPeer(wgp *wgtypes.Peer, i *BaseInterface) (*Peer, error) {
	var err error

	logger := zap.L().Named("peer").With(
		zap.String("intf", i.Name()),
		zap.Any("peer", wgp.PublicKey),
	)

	p := &Peer{
		Interface: i,
		Peer:      *wgp,
		client:    i.client,
		backend:   i.backend,
		events:    i.events,
		config:    i.config,
		doRestart: make(chan interface{}),
		logger:    logger,
	}

	// Setup proxy
	if p.proxy, err = proxy.NewProxy(i.nat, i.ListenPort, i.config.Proxy.EBPF, i.config.Proxy.NFT); err != nil {
		return nil, fmt.Errorf("failed to setup proxy: %w", err)
	}

	ip, err := p.PublicKey().IPv6Address()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP address: %w", err)
	}

	// Add default link-local address as allowed IP
	ap := net.IPNet{
		IP:   ip.IP,
		Mask: net.CIDRMask(128, 128),
	}
	if err := p.addAllowedIP(ap); err != nil {
		return nil, fmt.Errorf("failed to add link-local IPv6 address to AllowedIPs")
	}

	// Initialize signaling channel
	kp := p.PublicPrivateKeyPair()
	p.logger.Info("Subscribe to messages from peer", zap.Any("kp", kp))
	if p.messages, err = p.backend.Subscribe(context.Background(), kp); err != nil {
		p.logger.Fatal("Failed to subscribe to offers", zap.Error(err))
	}

	// Initialize ICE agent
	if p.agent, err = p.newAgent(); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	p.description = p.newDescription()

	p.logger.Info("Starting gathering local ICE candidates")
	if err := p.agent.GatherCandidates(); err != nil {
		p.logger.Fatal("Failed to gather candidates", zap.Error(err))
	}

	go p.run()

	return p, nil
}

func (p *Peer) run() {
	for {
		select {
		case <-p.doRestart:
			if err := p.restart(); err != nil {
				p.logger.Fatal("Failed to restart agent", zap.Error(err))
			}

		case msg := <-p.messages:
			if err := p.onMessage(msg); err != nil {
				p.logger.Error("Failed to handle message",
					zap.Error(err),
					zap.Any("msg", msg))
			}
		}
	}
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

// Restart the ICE agent by creating a nee one
func (p *Peer) Restart() {
	p.doRestart <- nil
}

func (p *Peer) restart() error {
	var err error

	p.logger.Debug("Restarting ICE session")

	if err := p.agent.Close(); err != nil {
		return fmt.Errorf("failed to close agent: %w", err)
	}
	p.logger.Debug("Agent closed")

	if p.agent, err = p.newAgent(); err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	p.description = p.newDescription()

	p.logger.Info("Starting gathering local ICE candidates")
	if err := p.agent.GatherCandidates(); err != nil {
		return fmt.Errorf("failed to gather candidates: %w", err)
	}

	return nil
}

func (p *Peer) sendDescription() error {
	p.description.Epoch++

	msg := &pb.SignalingMessage{
		Description: p.description,
	}

	if err := p.backend.Publish(context.Background(), p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Send session description", zap.Any("description", p.description))

	return nil
}

func (p *Peer) newAgent() (*ice.Agent, error) {
	var agent *ice.Agent
	var err error

	// Setup new ICE Agent
	config, err := p.config.AgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ICE agent: %w", err)
	}

	// Do not use Wireguard interfaces for ICE
	origFilter := config.InterfaceFilter
	config.InterfaceFilter = func(name string) bool {
		_, err := p.client.Device(name)
		return origFilter(name) && err != nil
	}

	config.UDPMux = p.Interface.udpMux
	config.UDPMuxSrflx = p.Interface.udpMuxSrflx
	config.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	if agent, err = ice.NewAgent(config); err != nil {
		return nil, fmt.Errorf("failed to create ICE agent: %w", err)
	}

	// When we have gathered a new ICE Candidate send it to the remote peer
	if err := agent.OnCandidate(p.onCandidate); err != nil {
		return nil, fmt.Errorf("failed to setup on candidate handler: %w", err)
	}

	// When selected candidate pair changes
	if err := agent.OnSelectedCandidatePairChange(p.onSelectedCandidatePairChange); err != nil {
		return nil, fmt.Errorf("failed to setup on selected candidate pair handler: %w", err)
	}

	// When ICE Connection state has change print to stdout
	if err := agent.OnConnectionStateChange(p.onConnectionStateChange); err != nil {
		return nil, fmt.Errorf("failed to setup on connection state handler: %w", err)
	}

	p.ConnectionState = ice.ConnectionStateNew
	p.logger.Debug("Agent created")

	return agent, nil
}

// newDescription initializes a new offer before it gets send via the signaling channel
func (p *Peer) newDescription() *pb.SessionDescription {
	ufrag, pwd, err := p.agent.GetLocalUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local user credentials", zap.Error(err))
	}

	return &pb.SessionDescription{
		Epoch:      0,
		Ufrag:      ufrag,
		Pwd:        pwd,
		Candidates: []*pb.Candidate{},
	}
}

// isControlling determines if the peer is controlling the ICE session
// by selecting the peer which has the smaller public key
func (p *Peer) isControlling() bool {
	var pkOur, pkTheir big.Int
	pkOur.SetBytes(p.Interface.Device.PublicKey[:])
	pkTheir.SetBytes(p.Peer.PublicKey[:])

	return pkOur.Cmp(&pkTheir) == -1
}

// onConnectionStateChange is a callback which gets called by the ICE agent
// whenever the state of the ICE connection has changed
func (p *Peer) onConnectionStateChange(state ice.ConnectionState) {
	p.ConnectionState = state

	p.logger.Info("Connection state changed",
		zap.String("state", strings.ToLower(state.String())))

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
		p.Restart()
	}
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

	if err := p.client.ConfigureDevice(p.Interface.Device.Name, cfg); err != nil {
		return fmt.Errorf("failed to update peer endpoint: %w", err)
	}

	p.logger.Debug("Peer endpoint updated", zap.Any("endpoint", addr))

	c := exec.Command("wg")
	o, _ := c.CombinedOutput()
	fmt.Println(string(o))

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

	credsChanged := (ufrag != "" && pwd != "") && (ufrag != o.Ufrag || pwd != o.Pwd)

	return p.conn != nil && credsChanged
}

func (p *Peer) addCandidates(sd *pb.SessionDescription) error {
	for _, c := range sd.Candidates {
		ic, err := c.ICECandidate()
		if err != nil {
			p.logger.Error("Failed to decode description. Ignoring...", zap.Error(err))
			continue
		}

		if err := p.agent.AddRemoteCandidate(ic); err != nil {
			return fmt.Errorf("failed to add remote candidate: %w", err)
		}

		p.logger.Debug("Add remote candidate", zap.Any("candidate", c))
	}

	return nil
}

func (p *Peer) connect(ufrag, pwd string) error {
	var err error

	p.ConnectionState = ConnectionStateConnecting
	if p.isControlling() {
		p.logger.Debug("Dialing...")
		p.conn, err = p.agent.Dial(context.Background(), ufrag, pwd)
	} else {
		p.logger.Debug("Accepting...")
		p.conn, err = p.agent.Accept(context.Background(), ufrag, pwd)
	}
	if err != nil {
		return err
	}

	p.logger.Debug("Connection established",
		zap.Any("localAddress", p.conn.LocalAddr()),
		zap.Any("remoteAddress", p.conn.RemoteAddr()))

	cp, err := p.agent.GetSelectedCandidatePair()
	if err != nil {
		return fmt.Errorf("failed to get selected candidate pair: %w", err)
	}

	ep, err := p.proxy.Update(cp, p.conn)
	if err != nil {
		return fmt.Errorf("failed to update proxy: %w", err)
	}

	if err := p.updateEndpoint(ep); err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}
