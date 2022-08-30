package epdisc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/feat/epdisc/proxy"
	"riasc.eu/wice/pkg/log"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/util"

	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
	proto "riasc.eu/wice/pkg/proto"
	protoepdisc "riasc.eu/wice/pkg/proto/feat/epdisc"
)

type Peer struct {
	*core.Peer

	Interface *Interface

	config          *config.Config
	agentConfig     *ice.AgentConfig
	agent           *ice.Agent
	backend         signaling.Backend
	proxy           proxy.Proxy
	connectionState util.AtomicEnum[icex.ConnectionState]
	lastStateChange time.Time
	lastEndpoint    *net.UDPAddr
	restarts        uint
	credentials     protoepdisc.Credentials

	signalingMessages      chan *signaling.Message
	connectionStateChanges chan icex.ConnectionState

	logger *zap.Logger
}

func NewPeer(cp *core.Peer, i *Interface) (*Peer, error) {
	var err error

	p := &Peer{
		Peer:      cp,
		Interface: i,

		backend: i.Discovery.backend,
		config:  i.Discovery.config,

		signalingMessages:      make(chan *signaling.Message, 32),
		connectionStateChanges: make(chan icex.ConnectionState, 32),

		logger: zap.L().Named("ice.peer").With(
			zap.String("intf", i.Name()),
			zap.Any("peer", cp.PublicKey()),
		),
	}

	p.connectionState.Store(ice.ConnectionStateClosed)

	// Initialize signaling channel
	kp := cp.PublicPrivateKeyPair()
	if _, err := p.backend.Subscribe(context.Background(), kp, p); err != nil {
		// TODO: Attempt retry?
		return nil, fmt.Errorf("failed to subscribe to offers: %w", err)
	}
	p.logger.Info("Subscribed to messages from peer", zap.Any("kp", kp))

	// Prepare ICE agent configuration
	p.agentConfig, err = p.config.AgentConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ICE agent configuration: %w", err)
	}

	// Do not use WireGuard interfaces for ICE
	origFilter := p.agentConfig.InterfaceFilter
	p.agentConfig.InterfaceFilter = func(name string) bool {
		return origFilter(name) && i.Discovery.watcher.InterfaceByName(name) == nil
	}

	p.agentConfig.UDPMux = i.udpMux
	p.agentConfig.UDPMuxSrflx = i.udpMuxSrflx
	p.agentConfig.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	// Setup proxy
	if dev, ok := i.KernelDevice.(*device.UserDevice); ok {
		if p.proxy, err = proxy.NewUserBindProxy(dev.Bind); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else if i.nat != nil {
		if p.proxy, err = proxy.NewKernelProxy(i.nat, cp.Interface.ListenPort); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else {
		return nil, fmt.Errorf("failed tp setup peer. Neither NAT or Bind is configured")
	}

	if err = p.createAgent(); err != nil {
		return nil, fmt.Errorf("failed to create initial agent: %w", err)
	}

	go p.run()

	return p, nil
}

func (p *Peer) Resubscribe(ctx context.Context, skOld crypto.Key) error {
	// TODO: Attempt retries?

	// Create new subscription
	kpNew := p.PublicPrivateKeyPair()
	if _, err := p.backend.Subscribe(ctx, kpNew, p); err != nil {
		return fmt.Errorf("failed to subscribe to offers: %w", err)
	}

	// Remove old subscription
	kpOld := &crypto.KeyPair{
		Ours:   skOld,
		Theirs: p.PublicKey(),
	}

	if _, err := p.backend.Unsubscribe(ctx, kpOld, p); err != nil {
		return fmt.Errorf("failed to unsubscribe from offers: %w", err)
	}

	p.logger.Info("Updated subcription",
		zap.Any("old", kpOld.Public()),
		zap.Any("new", kpNew.Public()))

	return nil
}

func (p *Peer) ConnectionState() icex.ConnectionState {
	return p.connectionState.Load()
}

func (p *Peer) run() {
	for {
		select {
		case msg := <-p.signalingMessages:
			p.onSignalingMessage(msg)

		case sc := <-p.connectionStateChanges:
			p.onConnectionStateChange(sc)
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

	p.restarts++

	return nil
}

func (p *Peer) sendCredentials(need bool) error {
	p.credentials.NeedCreds = need

	msg := &signaling.Message{
		Credentials: &p.credentials,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := p.backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent credentials", zap.Any("creds", msg.Credentials))

	return nil
}

func (p *Peer) sendCandidate(c ice.Candidate) error {
	msg := &signaling.Message{
		Candidate: protoepdisc.NewCandidate(c),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := p.backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent candidate", zap.Any("candidate", msg.Candidate))

	return nil
}

func (p *Peer) createAgent() error {
	var err error

	if !p.setConnectionStateIf(ice.ConnectionStateClosed, icex.ConnectionStateCreating) {
		return fmt.Errorf("failed to create new agent if previous one is not closed")
	}

	p.logger.Info("Creating new agent")

	p.credentials = protoepdisc.NewCredentials()

	p.agentConfig.LocalUfrag = p.credentials.Ufrag
	p.agentConfig.LocalPwd = p.credentials.Pwd

	// Setup new ICE Agent
	if p.agent, err = ice.NewAgent(p.agentConfig); err != nil {
		return fmt.Errorf("failed to create ICE agent: %w", err)
	}

	// When we have gathered a new ICE Candidate send it to the remote peer
	if err := p.agent.OnCandidate(p.onCandidate); err != nil {
		return fmt.Errorf("failed to setup on candidate handler: %w", err)
	}

	// When selected candidate pair changes
	if err := p.agent.OnSelectedCandidatePairChange(p.onSelectedCandidatePairChange); err != nil {
		return fmt.Errorf("failed to setup on selected candidate pair handler: %w", err)
	}

	// When ICE Connection state has change print to stdout
	if err := p.agent.OnConnectionStateChange(func(cs ice.ConnectionState) {
		p.onConnectionStateChange(icex.ConnectionState(cs))
	}); err != nil {
		return fmt.Errorf("failed to setup on connection state handler: %w", err)
	}

	if !p.setConnectionStateIf(icex.ConnectionStateCreating, icex.ConnectionStateIdle) {
		return fmt.Errorf("failed to switch to idle state")
	}

	if err := p.sendCredentials(true); err != nil {
		return fmt.Errorf("failed to send peer credentials: %w", err)
	}

	return nil
}

// isSessionRestart checks if a received offer should restart the
// ICE session by comparing ufrag & pwd with previously used values.
func (p *Peer) isSessionRestart(c *protoepdisc.Credentials) bool {
	ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local credentials", zap.Error(err))
	}

	credsChanged := (ufrag != "" && pwd != "") && (c.Ufrag != "" && c.Pwd != "") && (ufrag != c.Ufrag || pwd != c.Pwd)

	return p.ConnectionState() != ice.ConnectionStateClosed && credsChanged
}

func (p *Peer) addRemoteCandidate(c *protoepdisc.Candidate) error {
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
	var conn *ice.Conn

	// TODO: use proper context
	if p.IsControlling() {
		p.logger.Debug("Dialing...")
		conn, err = p.agent.Dial(context.Background(), ufrag, pwd)
	} else {
		p.logger.Debug("Accepting...")
		conn, err = p.agent.Accept(context.Background(), ufrag, pwd)
	}
	if err != nil {
		return err
	}

	cp, err := p.agent.GetSelectedCandidatePair()
	if err != nil {
		return fmt.Errorf("failed to get selected candidate pair: %w", err)
	}

	ep, err := p.proxy.UpdateCandidatePair(cp, conn)
	if err != nil {
		return fmt.Errorf("failed to update proxy: %w", err)
	}

	if err := p.UpdateEndpoint(ep); err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	p.lastEndpoint = ep

	return nil
}

func (p *Peer) setConnectionState(new icex.ConnectionState) icex.ConnectionState {
	prev := p.connectionState.Swap(new)

	p.lastStateChange = time.Now()

	p.logger.Info("Connection state changed",
		zap.String("new", strings.ToLower(new.String())),
		zap.String("previous", strings.ToLower(prev.String())))

	for _, h := range p.Interface.Discovery.onConnectionStateChange {
		h.OnConnectionStateChange(p, new, prev)
	}

	return prev
}

func (p *Peer) setConnectionStateIf(prev, new icex.ConnectionState) bool {
	swapped := p.connectionState.CompareAndSwap(prev, new)
	if swapped {
		p.lastStateChange = time.Now()

		p.logger.Info("Connection state changed",
			zap.String("new", strings.ToLower(new.String())),
			zap.String("previous", strings.ToLower(prev.String())))

		for _, h := range p.Interface.Discovery.onConnectionStateChange {
			h.OnConnectionStateChange(p, new, prev)
		}
	}

	return swapped
}

func (p *Peer) Marshal() *protoepdisc.Peer {
	cs := p.ConnectionState()

	q := &protoepdisc.Peer{
		State:        protoepdisc.NewConnectionState(cs),
		Restarts:     uint32(p.restarts),
		Reachability: p.Reachability(),
	}

	if p.proxy != nil {
		q.ProxyType = p.proxy.Type()
	}

	if !p.lastStateChange.IsZero() {
		q.LastStateChangeTimestamp = proto.Time(p.lastStateChange)
	}

	if cs != ice.ConnectionStateClosed {
		cp, err := p.agent.GetSelectedCandidatePair()
		if err == nil && cp != nil {
			q.SelectedCandidatePair = &protoepdisc.CandidatePair{
				Local:  protoepdisc.NewCandidate(cp.Local),
				Remote: protoepdisc.NewCandidate(cp.Remote),
			}
		}

		for _, cps := range p.agent.GetCandidatePairsStats() {
			q.CandidatePairStats = append(q.CandidatePairStats, protoepdisc.NewCandidatePairStats(&cps))
		}

		for _, cs := range p.agent.GetLocalCandidatesStats() {
			q.LocalCandidateStats = append(q.LocalCandidateStats, protoepdisc.NewCandidateStats(&cs))
		}

		for _, cs := range p.agent.GetRemoteCandidatesStats() {
			q.RemoteCandidateStats = append(q.RemoteCandidateStats, protoepdisc.NewCandidateStats(&cs))
		}
	}

	return q
}

func (p *Peer) Reachability() protoepdisc.Reachability {
	switch p.ConnectionState() {
	case ice.ConnectionStateConnected:
		cp, err := p.agent.GetSelectedCandidatePair()
		if err != nil {
			return protoepdisc.Reachability_NO_REACHABILITY
		}

		switch cp.Remote.Type() {
		case ice.CandidateTypeHost:
			fallthrough
		case ice.CandidateTypeServerReflexive:
			if cp.Remote.NetworkType().IsTCP() {
				return protoepdisc.Reachability_DIRECT_TCP
			} else {
				return protoepdisc.Reachability_DIRECT_UDP
			}

		case ice.CandidateTypeRelay:
			if cp.Remote.NetworkType().IsTCP() {
				return protoepdisc.Reachability_RELAY_TCP
			} else {
				return protoepdisc.Reachability_RELAY_UDP
			}
		}
	}

	return protoepdisc.Reachability_NO_REACHABILITY
}
