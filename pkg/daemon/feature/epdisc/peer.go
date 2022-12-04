package epdisc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pion/ice/v2"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc/proxy"
	"github.com/stv0g/cunicu/pkg/device"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/pkg/log"
	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"
	"go.uber.org/zap"
)

var (
	errNoNATorBind          = errors.New("failed tp setup peer. Neither NAT or Bind is configured")
	errCreateNonClosedAgent = errors.New("failed to create new agent if previous one is not closed")
	errSwitchToIdle         = errors.New("failed to switch to idle state")
	errStillIdle            = errors.New("not connected yet")
)

type Peer struct {
	*core.Peer
	Interface *Interface

	agent                  *ice.Agent
	proxy                  proxy.Proxy
	connectionState        util.AtomicEnum[icex.ConnectionState]
	lastStateChange        time.Time
	lastEndpoint           *net.UDPAddr
	restarts               uint
	credentials            epdiscproto.Credentials
	signalingMessages      chan *signaling.Message
	connectionStateChanges chan icex.ConnectionState

	logger *zap.Logger
}

func NewPeer(cp *core.Peer, e *Interface) (*Peer, error) {
	var err error

	p := &Peer{
		Peer:      cp,
		Interface: e,

		signalingMessages:      make(chan *signaling.Message, 32),
		connectionStateChanges: make(chan icex.ConnectionState, 32),

		logger: e.logger.Named("peer").With(
			zap.String("peer", cp.String()),
		),
	}

	p.connectionState.Store(ice.ConnectionStateClosed)

	// Initialize signaling channel
	kp := p.PublicPrivateKeyPair()
	if _, err := p.Interface.Daemon.Backend.Subscribe(context.Background(), kp, p); err != nil {
		// TODO: Attempt retry?
		return nil, fmt.Errorf("failed to subscribe to offers: %w", err)
	}
	p.logger.Info("Subscribed to messages from peer", zap.Any("kp", kp))

	// Setup proxy
	//nolint:gocritic
	if dev, ok := e.KernelDevice.(*device.UserDevice); ok {
		if p.proxy, err = proxy.NewUserBindProxy(dev.Bind); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else if e.nat != nil {
		if p.proxy, err = proxy.NewKernelProxy(e.nat, cp.Interface.ListenPort); err != nil {
			return nil, fmt.Errorf("failed to setup proxy: %w", err)
		}
	} else {
		return nil, errNoNATorBind
	}

	if err = p.createAgentWithBackoff(); err != nil {
		return nil, fmt.Errorf("failed to create initial agent: %w", err)
	}

	go p.run()

	return p, nil
}

func (p *Peer) Resubscribe(ctx context.Context, skOld crypto.Key) error {
	// Create new subscription
	kpNew := p.PublicPrivateKeyPair()
	if _, err := p.Interface.Daemon.Backend.Subscribe(ctx, kpNew, p); err != nil {
		return fmt.Errorf("failed to subscribe to offers: %w", err)
	}

	// Remove old subscription
	kpOld := &crypto.KeyPair{
		Ours:   skOld,
		Theirs: p.PublicKey(),
	}

	if _, err := p.Interface.Daemon.Backend.Unsubscribe(ctx, kpOld, p); err != nil {
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

	if err := p.Interface.Daemon.Backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent credentials", zap.Any("creds", msg.Credentials))

	return nil
}

func (p *Peer) sendCandidate(c ice.Candidate) error {
	msg := &signaling.Message{
		Candidate: epdiscproto.NewCandidate(c),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := p.Interface.Daemon.Backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent candidate", zap.Any("candidate", msg.Candidate))

	return nil
}

func (p *Peer) createAgentWithBackoff() error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 1 * time.Minute

	return backoff.RetryNotify(
		func() error {
			return p.createAgent()
		}, bo,
		func(err error, d time.Duration) {
			p.logger.Error("Failed to create agent",
				zap.Error(err),
				zap.Duration("after", d))
		},
	)
}

func (p *Peer) createAgent() error {
	var err error

	if !p.setConnectionStateIf(ice.ConnectionStateClosed, icex.ConnectionStateCreating) {
		return errCreateNonClosedAgent
	}

	// Reset state to closed if we error-out of this function
	defer func() {
		p.setConnectionStateIf(icex.ConnectionStateCreating, ice.ConnectionStateClosed)
	}()

	p.logger.Info("Creating new agent")

	// Prepare ICE agent configuration
	pk := p.Interface.PublicKey()
	acfg, err := p.Interface.Settings.AgentConfig(context.Background(), &pk)
	if err != nil {
		return fmt.Errorf("failed to generate ICE agent configuration: %w", err)
	}

	// Do not use WireGuard interfaces for ICE
	origFilter := acfg.InterfaceFilter
	acfg.InterfaceFilter = func(name string) bool {
		return origFilter(name) && p.Interface.Daemon.InterfaceByName(name) == nil
	}

	acfg.UDPMux = p.Interface.udpMux
	acfg.UDPMuxSrflx = p.Interface.udpMuxSrflx
	acfg.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	p.credentials = epdiscproto.NewCredentials()

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
		return errSwitchToIdle
	}

	// Send peer credentials as long as we remain in ConnectionStateIdle
	go func() {
		if err := p.sendCredentialsWithBackoff(true); err != nil {
			p.logger.Error("Failed to send credentials", zap.Error(err))
		}
	}()

	return nil
}

func (p *Peer) sendCredentialsWithBackoff(need bool) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 1 * time.Minute

	return backoff.RetryNotify(
		func() error {
			if err := p.sendCredentials(need); err != nil {
				return err
			}

			if p.ConnectionState() == icex.ConnectionStateIdle {
				return errStillIdle
			}

			return nil
		}, bo,
		func(err error, d time.Duration) {
			if errors.Is(err, errStillIdle) {
				p.logger.Error("Failed to send peer credentials",
					zap.Error(err),
					zap.Duration("after", d))
			}
		},
	)
}

// isSessionRestart checks if a received offer should restart the
// ICE session by comparing ufrag & pwd with previously used values.
func (p *Peer) isSessionRestart(c *epdiscproto.Credentials) bool {
	ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
	if err != nil {
		p.logger.Error("Failed to get local credentials", zap.Error(err))
	}

	credsChanged := (ufrag != "" && pwd != "") && (c.Ufrag != "" && c.Pwd != "") && (ufrag != c.Ufrag || pwd != c.Pwd)

	return p.ConnectionState() != ice.ConnectionStateClosed && credsChanged
}

func (p *Peer) addRemoteCandidate(c *epdiscproto.Candidate) error {
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

	if err := p.SetEndpoint(ep); err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	p.lastEndpoint = ep

	return nil
}

// setConnectionState updates the connection state of the peer and invokes registered handlers.
// It returns the previous connection state.
func (p *Peer) setConnectionState(newState icex.ConnectionState) icex.ConnectionState { //nolint:unparam
	prevState := p.connectionState.Swap(newState)

	p.lastStateChange = time.Now()

	p.logger.Info("Connection state changed",
		zap.String("new", strings.ToLower(newState.String())),
		zap.String("previous", strings.ToLower(prevState.String())))

	for _, h := range p.Interface.onConnectionStateChange {
		h.OnConnectionStateChange(p, newState, prevState)
	}

	return prevState
}

// setConnectionStateIf updates the connection state of the peer if the previous state matches the one supplied.
// It returns true if the state has been changed.
func (p *Peer) setConnectionStateIf(prevState, newState icex.ConnectionState) bool {
	swapped := p.connectionState.CompareAndSwap(prevState, newState)
	if swapped {
		p.lastStateChange = time.Now()

		p.logger.Info("Connection state changed",
			zap.String("new", strings.ToLower(newState.String())),
			zap.String("previous", strings.ToLower(prevState.String())))

		for _, h := range p.Interface.onConnectionStateChange {
			h.OnConnectionStateChange(p, newState, prevState)
		}
	}

	return swapped
}

// Marshal marshals a description of the peer into a Protobuf description
func (p *Peer) Marshal() *epdiscproto.Peer {
	cs := p.ConnectionState()

	q := &epdiscproto.Peer{
		State:    epdiscproto.NewConnectionState(cs),
		Restarts: uint32(p.restarts),
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
			q.SelectedCandidatePair = &epdiscproto.CandidatePair{
				Local:  epdiscproto.NewCandidate(cp.Local),
				Remote: epdiscproto.NewCandidate(cp.Remote),
			}
		}

		for _, cps := range p.agent.GetCandidatePairsStats() {
			cps := cps
			q.CandidatePairStats = append(q.CandidatePairStats, epdiscproto.NewCandidatePairStats(&cps))
		}

		for _, cs := range p.agent.GetLocalCandidatesStats() {
			cs := cs
			q.LocalCandidateStats = append(q.LocalCandidateStats, epdiscproto.NewCandidateStats(&cs))
		}

		for _, cs := range p.agent.GetRemoteCandidatesStats() {
			cs := cs
			q.RemoteCandidateStats = append(q.RemoteCandidateStats, epdiscproto.NewCandidateStats(&cs))
		}
	}

	return q
}

func (p *Peer) Reachability() coreproto.ReachabilityType {
	cs := p.connectionState.Load()
	switch cs {
	case icex.ConnectionStateConnecting,
		icex.ConnectionStateCreating,
		icex.ConnectionStateIdle,
		ice.ConnectionStateChecking,
		ice.ConnectionStateNew:
		return coreproto.ReachabilityType_REACHABILITY_TYPE_UNKNOWN

	case ice.ConnectionStateClosed,
		ice.ConnectionStateDisconnected,
		ice.ConnectionStateFailed:
		return coreproto.ReachabilityType_REACHABILITY_TYPE_NONE

	case ice.ConnectionStateConnected:
		cp, err := p.agent.GetSelectedCandidatePair()
		if err != nil || cp == nil {
			return coreproto.ReachabilityType_REACHABILITY_TYPE_NONE
		}

		lc, rc := cp.Local, cp.Remote

		switch {
		case lc.Type() == ice.CandidateTypeRelay && rc.Type() == ice.CandidateTypeRelay:
			return coreproto.ReachabilityType_REACHABILITY_TYPE_RELAYED_BIDIR
		case lc.Type() == ice.CandidateTypeRelay || rc.Type() == ice.CandidateTypeRelay:
			return coreproto.ReachabilityType_REACHABILITY_TYPE_RELAYED
		default:
			return coreproto.ReachabilityType_REACHABILITY_TYPE_DIRECT
		}

	default:
		return coreproto.ReachabilityType_REACHABILITY_TYPE_NONE
	}
}
