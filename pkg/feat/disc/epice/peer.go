package epice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/device"
	icex "riasc.eu/wice/pkg/ice"
	"riasc.eu/wice/pkg/log"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/proxy"
	"riasc.eu/wice/pkg/signaling"
)

type Peer struct {
	*core.Peer

	Interface *Interface

	config *config.Config

	backend signaling.Backend
	proxy   proxy.Proxy

	// TODO: Avoid races around connection state
	ConnectionState icex.ConnectionState

	agentConfig *ice.AgentConfig
	agent       *ice.Agent
	conn        *ice.Conn
	credentials pb.Credentials

	logger *zap.Logger
}

func NewPeer(cp *core.Peer, i *Interface) (*Peer, error) {
	var err error

	p := &Peer{
		Peer:            cp,
		Interface:       i,
		ConnectionState: icex.ConnectionStateUnknown,

		backend: i.Discovery.backend,
		config:  i.Discovery.config,

		logger: zap.L().Named("ice.peer").With(
			zap.String("intf", i.Name()),
			zap.Any("peer", cp.PublicKey()),
		),
	}

	// Prepare ICE agent configuration
	p.agentConfig, err = p.config.AgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ICE agent configuration: %w", err)
	}

	// Do not use WireGuard interfaces for ICE
	origFilter := p.agentConfig.InterfaceFilter
	p.agentConfig.InterfaceFilter = func(name string) bool {
		return origFilter(name) && i.Discovery.watcher.Interfaces.ByName(name) == nil
	}

	p.agentConfig.UDPMux = i.udpMux
	p.agentConfig.UDPMuxSrflx = i.udpMuxSrflx
	p.agentConfig.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	// Setup proxy
	if dev, ok := i.KernelDevice.(*device.UserDevice); ok {
		if p.proxy, err = proxy.NewUserProxy(dev.Bind); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else if i.nat != nil {
		if p.proxy, err = proxy.NewKernelProxy(i.nat, cp.Interface.ListenPort); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed tp setup peer. Neither NAT or Bind is configured")
	}

	// Initialize signaling channel
	kp := p.PublicPrivateKeyPair()
	if err := p.backend.Subscribe(context.Background(), kp, p); err != nil {
		p.logger.Fatal("Failed to subscribe to offers", zap.Error(err))
	}

	p.logger.Info("Subscribed to messages from peer", zap.Any("kp", kp))

	// Initialize new agent by simulating a closed event
	p.onConnectionStateChange(ice.ConnectionStateClosed)

	return p, nil
}

// Close destroys the peer as well as the ICE agent and proxies
func (p *Peer) Close() error {
	p.setConnectionState(icex.ConnectionStateClosing)

	if err := p.agent.Close(); err != nil {
		return fmt.Errorf("failed to close ICE agent: %w", err)
	}

	if err := p.proxy.Close(); err != nil {
		return fmt.Errorf("failed to close proxy: %w", err)
	}

	return nil
}

// Restart the ICE agent by creating a new one
func (p *Peer) Restart() error {
	p.logger.Debug("Restarting ICE session")

	if err := p.agent.Close(); err != nil {
		return fmt.Errorf("failed to close agent: %w", err)
	}

	// The new agent will be recreated in the onConnectionStateChange() handler
	// once the old agent has been properly closed

	return nil
}

func (p *Peer) sendCredentials(need bool) error {
	p.credentials.NeedCreds = need

	msg := &signaling.Message{
		Credentials: &p.credentials,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := p.backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent credentials", zap.Any("creds", msg.Credentials))

	return nil
}

func (p *Peer) sendCandidate(c ice.Candidate) error {
	msg := &signaling.Message{
		Candidate: pb.NewCandidate(c),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := p.backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent candidate", zap.Any("candidate", msg.Candidate))

	return nil
}

func (p *Peer) newAgent() (*ice.Agent, error) {
	var agent *ice.Agent
	var err error

	p.logger.Info("Creating new agent")

	p.credentials = pb.NewCredentials()

	p.agentConfig.LocalUfrag = p.credentials.Ufrag
	p.agentConfig.LocalPwd = p.credentials.Pwd

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

	p.setConnectionState(icex.ConnectionStateIdle)

	return agent, nil
}

// isSessionRestart checks if a received offer should restart the
// ICE session by comparing ufrag & pwd with previously used values.
func (p *Peer) isSessionRestart(c *pb.Credentials) bool {
	ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local credentials", zap.Error(err))
	}

	credsChanged := (ufrag != "" && pwd != "") && (c.Ufrag != "" && c.Pwd != "") && (ufrag != c.Ufrag || pwd != c.Pwd)

	return p.conn != nil && credsChanged
}

func (p *Peer) addRemoteCandidate(c *pb.Candidate) error {
	ic, err := c.ICECandidate()
	if err != nil {
		return fmt.Errorf("failed to remote candidate: %w", err)
	}

	if err := p.agent.AddRemoteCandidate(ic); err != nil {
		return fmt.Errorf("failed to add remote candidate: %w", err)
	}

	p.logger.Debug("Add remote candidate", zap.Any("candidate", c))

	return nil
}

func (p *Peer) connect(ufrag, pwd string) error {
	var err error

	// TODO: use proper context
	if p.IsControlling() {
		p.logger.Debug("Dialing...")
		p.conn, err = p.agent.Dial(context.Background(), ufrag, pwd)
	} else {
		p.logger.Debug("Accepting...")
		p.conn, err = p.agent.Accept(context.Background(), ufrag, pwd)
	}
	if err != nil {
		return err
	}

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

func (p *Peer) setConnectionState(new icex.ConnectionState) icex.ConnectionState {
	prev := p.ConnectionState
	p.ConnectionState = new

	// Suppress initial invocation of handler from NewPeer()
	if prev != icex.ConnectionStateUnknown {
		p.logger.Info("Connection state changed",
			zap.String("new", strings.ToLower(new.String())),
			zap.String("previous", strings.ToLower(prev.String())))

		for _, h := range p.Interface.Discovery.onConnectionStateChange {
			h.OnConnectionStateChange(p, new, prev)
		}
	}

	return prev
}
