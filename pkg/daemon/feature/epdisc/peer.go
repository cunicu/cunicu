package epdisc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jpillora/backoff"
	"github.com/pion/ice/v2"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc/proxy"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"

	icex "github.com/stv0g/cunicu/pkg/ice"
	proto "github.com/stv0g/cunicu/pkg/proto"
	protoepdisc "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

type Peer struct {
	*core.Peer

	intf                   *Interface
	agent                  *ice.Agent
	proxy                  proxy.Proxy
	connectionState        util.AtomicEnum[icex.ConnectionState]
	lastStateChange        time.Time
	lastEndpoint           *net.UDPAddr
	restarts               uint
	credentials            protoepdisc.Credentials
	signalingMessages      chan *signaling.Message
	connectionStateChanges chan icex.ConnectionState

	logger *zap.Logger
}

func NewPeer(cp *core.Peer, e *Interface) (*Peer, error) {
	var err error

	p := &Peer{
		Peer: cp,
		intf: e,

		signalingMessages:      make(chan *signaling.Message, 32),
		connectionStateChanges: make(chan icex.ConnectionState, 32),

		logger: e.logger.Named("peer").With(
			zap.String("peer", cp.String()),
		),
	}

	p.connectionState.Store(ice.ConnectionStateClosed)

	// Initialize signaling channel
	kp := p.PublicPrivateKeyPair()
	if _, err := p.intf.Daemon.Backend.Subscribe(context.Background(), kp, p); err != nil {
		// TODO: Attempt retry?
		return nil, fmt.Errorf("failed to subscribe to offers: %w", err)
	}
	p.logger.Info("Subscribed to messages from peer", zap.Any("kp", kp))

	// Setup proxy
	if dev, ok := e.KernelDevice.(*device.UserDevice); ok {
		if p.proxy, err = proxy.NewUserBindProxy(dev.Bind); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else if e.nat != nil {
		if p.proxy, err = proxy.NewKernelProxy(e.nat, cp.Interface.ListenPort); err != nil {
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
	// Create new subscription
	kpNew := p.PublicPrivateKeyPair()
	if _, err := p.intf.Daemon.Backend.Subscribe(ctx, kpNew, p); err != nil {
		return fmt.Errorf("failed to subscribe to offers: %w", err)
	}

	// Remove old subscription
	kpOld := &crypto.KeyPair{
		Ours:   skOld,
		Theirs: p.PublicKey(),
	}

	if _, err := p.intf.Daemon.Backend.Unsubscribe(ctx, kpOld, p); err != nil {
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

	if err := p.intf.Daemon.Backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
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

	if err := p.intf.Daemon.Backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
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

	// Prepare ICE agent configuration
	pk := p.Interface.PublicKey()
	acfg, err := p.intf.Settings.AgentConfig(context.Background(), &pk)
	if err != nil {
		return fmt.Errorf("failed to generate ICE agent configuration: %w", err)
	}

	// Do not use WireGuard interfaces for ICE
	origFilter := acfg.InterfaceFilter
	acfg.InterfaceFilter = func(name string) bool {
		return origFilter(name) && p.intf.Daemon.InterfaceByName(name) == nil
	}

	acfg.UDPMux = p.intf.udpMux
	acfg.UDPMuxSrflx = p.intf.udpMuxSrflx
	acfg.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	p.credentials = protoepdisc.NewCredentials()

	acfg.LocalUfrag = p.credentials.Ufrag
	acfg.LocalPwd = p.credentials.Pwd

	// Setup new ICE Agent
	if p.agent, err = ice.NewAgent(acfg); err != nil {
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

	// Send peer credentials as long as we remain in ConnectionStateIdle
	go p.resendCredentialsWithBackoff(true)

	return nil
}

func (p *Peer) resendCredentialsWithBackoff(need bool) {
	bo := backoff.Backoff{
		Factor: 1.6,
		Min:    time.Second,
		Max:    time.Minute,
	}

	for p.ConnectionState() == icex.ConnectionStateIdle {
		if err := p.sendCredentials(need); err != nil {
			p.logger.Error("Failed to send peer credentials", zap.Error(err))
		}

		time.Sleep(bo.Duration())
	}
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

// setConnectionState updates the connection state of the peer and invokes registered handlers.
// It returns the previous connection state.
func (p *Peer) setConnectionState(new icex.ConnectionState) icex.ConnectionState {
	prev := p.connectionState.Swap(new)

	p.lastStateChange = time.Now()

	p.logger.Info("Connection state changed",
		zap.String("new", strings.ToLower(new.String())),
		zap.String("previous", strings.ToLower(prev.String())))

	for _, h := range p.intf.onConnectionStateChange {
		h.OnConnectionStateChange(p, new, prev)
	}

	return prev
}

// setConnectionStateIf updates the connection state of the peer if the previous state matches the one supplied.
// It returns true if the state has been changed.
func (p *Peer) setConnectionStateIf(prev, new icex.ConnectionState) bool {
	swapped := p.connectionState.CompareAndSwap(prev, new)
	if swapped {
		p.lastStateChange = time.Now()

		p.logger.Info("Connection state changed",
			zap.String("new", strings.ToLower(new.String())),
			zap.String("previous", strings.ToLower(prev.String())))

		for _, h := range p.intf.onConnectionStateChange {
			h.OnConnectionStateChange(p, new, prev)
		}
	}

	return swapped
}

// Marshal marshals a description of the peer into a Protobuf description
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

	if p.agent != nil && cs != ice.ConnectionStateClosed {
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
