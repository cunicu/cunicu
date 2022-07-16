package ice

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/internal/log"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"
)

const (
	ConnectionStateConnecting = 100
)

type Peer struct {
	*core.Peer

	Interface *Interface

	config *config.Config

	backend signaling.Backend
	proxy   *proxy.Proxy

	ConnectionState ice.ConnectionState

	agentConfig *ice.AgentConfig
	agent       *ice.Agent
	conn        *ice.Conn

	description *pb.SessionDescription

	messages  chan *pb.SignalingMessage
	doRestart chan any

	logger *zap.Logger
}

func NewPeer(cp *core.Peer, i *Interface) (*Peer, error) {
	var err error

	p := &Peer{
		Peer:      cp,
		Interface: i,
		backend:   i.Discovery.backend,
		config:    i.Discovery.config,
		doRestart: make(chan any),

		logger: zap.L().Named("ice.peer"),
	}

	// Prepare ICE agent configuration
	p.agentConfig, err = p.config.AgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ICE agent configuration: %w", err)
	}

	// Do not use Wireguard interfaces for ICE
	origFilter := p.agentConfig.InterfaceFilter
	p.agentConfig.InterfaceFilter = func(name string) bool {
		return origFilter(name) && i.Discovery.watcher.Interfaces.ByName(name) == nil
	}

	p.agentConfig.UDPMux = i.udpMux
	p.agentConfig.UDPMuxSrflx = i.udpMuxSrflx
	p.agentConfig.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	// Setup proxy
	if p.proxy, err = proxy.NewProxy(i.nat, cp.Interface.ListenPort); err != nil {
		return nil, fmt.Errorf("failed to setup proxy: %w", err)
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

	if err := p.proxy.Close(); err != nil {
		return fmt.Errorf("failed to close proxy: %w", err)
	}

	p.logger.Info("Closed peer")

	return nil
}

// Restart the ICE agent by creating a new one
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
		Session: p.description,
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
	if agent, err = ice.NewAgent(p.agentConfig); err != nil {
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
	pkTheir.SetBytes(p.Peer.Peer.PublicKey[:])

	return pkOur.Cmp(&pkTheir) == -1
}

// onConnectionStateChange is a callback which gets called by the ICE agent
// whenever the state of the ICE connection has changed
func (p *Peer) onConnectionStateChange(cs ice.ConnectionState) {
	p.ConnectionState = cs

	p.logger.Info("Connection state changed",
		zap.String("state", strings.ToLower(cs.String())))

	p.Interface.Discovery.OnConnectionStateChange.Invoke(p, cs)

	if cs == ice.ConnectionStateFailed {
		p.Restart()
	}
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

	if err := p.UpdateEndpoint(ep); err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}
