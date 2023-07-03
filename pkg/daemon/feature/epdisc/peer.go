// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/types"
	"github.com/stv0g/cunicu/pkg/wg"
)

var (
	errCreateNonClosedAgent             = errors.New("failed to create new agent if previous one is not closed")
	errSwitchToIdle                     = errors.New("failed to switch to idle state")
	errStillIdle                        = errors.New("not connected yet")
	errClosing                          = errors.New("already closing")
	errInvalidConnectionStateForRestart = errors.New("can not restart agent while in state")
)

type Peer struct {
	*daemon.Peer
	Interface *Interface

	connectionState types.AtomicEnum[ConnectionState]

	agent    *ice.Agent
	proxy    Proxy
	endpoint *net.UDPAddr
	restarts atomic.Uint32

	remoteCredentials *epdiscproto.Credentials
	localCredentials  *epdiscproto.Credentials

	signalingMessages chan *signaling.Message
	newConnection     chan *ice.Conn

	logger *log.Logger
}

func NewPeer(cp *daemon.Peer, e *Interface) (*Peer, error) {
	p := &Peer{
		Peer:      cp,
		Interface: e,

		signalingMessages: make(chan *signaling.Message, 100),
		newConnection:     make(chan *ice.Conn),
		logger: e.logger.Named("peer").With(
			zap.String("peer", cp.String()),
		),
	}

	p.connectionState.Store(ConnectionStateClosed)

	e.Bind().AddOpenHandler(p)

	// Initialize signaling channel
	kp := p.PublicPrivateKeyPair()
	if _, err := p.Interface.Daemon.Backend.Subscribe(context.Background(), kp, p); err != nil {
		// TODO: Attempt retry?
		return nil, fmt.Errorf("failed to subscribe to offers: %w", err)
	}
	p.logger.Info("Subscribed to messages from peer", zap.Any("kp", kp))

	go p.createAgentWithBackoff()
	go p.run()

	return p, nil
}

// Getters

func (p *Peer) ConnectionState() ConnectionState {
	return p.connectionState.Load()
}

// Close destroys the peer as well as the ICE agent and proxies
func (p *Peer) Close() error {
	p.Interface.Bind().RemoveOpenHandler(p)

	if _, ok := p.connectionState.SetIfNot(ConnectionStateClosing, ConnectionStateClosing); !ok {
		return errClosing
	}

	kp := p.PublicPrivateKeyPair()
	if _, err := p.Interface.Daemon.Backend.Unsubscribe(context.Background(), kp, p); err != nil {
		return fmt.Errorf("failed to unsubscribe from offers: %w", err)
	}

	if p.agent != nil {
		if err := p.agent.Close(); err != nil && !errors.Is(err, ice.ErrClosed) {
			return fmt.Errorf("failed to close ICE agent: %w", err)
		}
	}

	if p.proxy != nil {
		if err := p.proxy.Close(); err != nil {
			return fmt.Errorf("failed to close proxy: %w", err)
		}
	}

	p.connectionState.SetIf(ConnectionStateClosed, ConnectionStateClosing)

	return nil
}

// Marshal marshals a description of the peer into a Protobuf description
func (p *Peer) Marshal() *epdiscproto.Peer {
	q := &epdiscproto.Peer{
		Restarts: p.restarts.Load(),
	}

	if p.proxy == nil {
		q.ProxyType = epdiscproto.ProxyType_NO_PROXY
	} else {
		switch p.proxy.(type) {
		case *BindProxy:
			q.ProxyType = epdiscproto.ProxyType_USER_BIND
		case *KernelConnProxy:
			q.ProxyType = epdiscproto.ProxyType_KERNEL_CONN
		case *KernelNATProxy:
			q.ProxyType = epdiscproto.ProxyType_KERNEL_NAT
		}
	}

	if !p.LastStateChangeTime.IsZero() {
		q.LastStateChangeTimestamp = proto.Time(p.LastStateChangeTime)
	}

	if p.agent != nil && p.State() != daemon.PeerStateClosed {
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
	switch p.ConnectionState() {
	case ConnectionStateConnecting,
		ConnectionStateCreating,
		ConnectionStateIdle,
		ConnectionStateChecking,
		ConnectionStateNew:
		return coreproto.ReachabilityType_UNSPECIFIED_REACHABILITY_TYPE

	case ConnectionStateClosed,
		ConnectionStateDisconnected,
		ConnectionStateFailed:
		return coreproto.ReachabilityType_NONE

	case ConnectionStateConnected:
		cp, err := p.agent.GetSelectedCandidatePair()
		if err != nil || cp == nil {
			return coreproto.ReachabilityType_NONE
		}

		lc, rc := cp.Local, cp.Remote

		switch {
		case lc.Type() == ice.CandidateTypeRelay && rc.Type() == ice.CandidateTypeRelay:
			return coreproto.ReachabilityType_RELAYED_BIDIR
		case lc.Type() == ice.CandidateTypeRelay || rc.Type() == ice.CandidateTypeRelay:
			return coreproto.ReachabilityType_RELAYED
		default:
			return coreproto.ReachabilityType_DIRECT
		}

	default:
		return coreproto.ReachabilityType_NONE
	}
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

// Restart the ICE agent by creating a new one
func (p *Peer) Restart() error {
	if prev, ok := p.connectionState.SetIfNot(ConnectionStateRestarting, ConnectionStateClosed, ConnectionStateClosing, ConnectionStateRestarting); !ok {
		return fmt.Errorf("%w: %s", errInvalidConnectionStateForRestart, strings.ToLower(prev.String()))
	}

	p.logger.Debug("Restarting ICE session")

	if err := p.agent.Close(); err != nil {
		return fmt.Errorf("failed to close agent: %w", err)
	}

	// The new agent will be recreated in the onConnectionStateChange() handler
	// once the old agent has been properly closed

	p.restarts.Add(1)

	return nil
}

func (p *Peer) run() {
	for msg := range p.signalingMessages {
		p.onSignalingMessage(msg)
	}
}

func (p *Peer) sendCredentials(need bool) error {
	p.localCredentials.NeedCreds = need

	msg := &signaling.Message{
		Credentials: p.localCredentials,
	}

	// TODO: Is this timeout suitable?
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := p.Interface.Daemon.Backend.Publish(ctx, p.PublicPrivateKeyPair(), msg); err != nil {
		return err
	}

	p.logger.Debug("Sent credentials", zap.Reflect("creds", msg.Credentials))

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

	p.logger.Debug("Sent candidate", zap.Reflect("candidate", msg.Candidate))

	return nil
}

func (p *Peer) createAgentWithBackoff() {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 1 * time.Minute

	if err := backoff.RetryNotify(
		func() error {
			return p.createAgent()
		}, bo,
		func(err error, d time.Duration) {
			p.logger.Error("Failed to create agent",
				zap.Error(err),
				zap.Duration("after", d))
		},
	); err != nil {
		p.logger.Error("Failed to create agent", zap.Error(err))
	}
}

func (p *Peer) createAgent() error {
	if _, ok := p.connectionState.SetIf(ConnectionStateCreating, ConnectionStateClosed); !ok {
		return errCreateNonClosedAgent
	}

	// Reset state to closed if we error-out of this function
	defer p.connectionState.SetIf(ConnectionStateClosed, ConnectionStateCreating)

	p.logger.Info("Creating new agent")

	// Prepare ICE agent configuration
	pk := p.Interface.PublicKey()
	acfg, err := p.Interface.Settings.AgentConfig(context.TODO(), &pk)
	if err != nil {
		return fmt.Errorf("failed to generate ICE agent configuration: %w", err)
	}

	// Do not use WireGuard interfaces for ICE
	origFilter := acfg.InterfaceFilter
	acfg.InterfaceFilter = func(name string) bool {
		return origFilter(name) && p.Interface.Daemon.InterfaceByName(name) == nil
	}

	acfg.UDPMux = p.Interface.mux
	acfg.UDPMuxSrflx = p.Interface.muxSrflx
	acfg.LoggerFactory = log.NewPionLoggerFactory(p.logger)

	p.localCredentials = epdiscproto.NewCredentials()
	p.remoteCredentials = nil

	acfg.LocalUfrag = p.localCredentials.Ufrag
	acfg.LocalPwd = p.localCredentials.Pwd

	// Setup new ICE Agent
	if p.agent, err = ice.NewAgent(acfg); err != nil {
		return fmt.Errorf("failed to create ICE agent: %w", err)
	}

	// When we have gathered a new ICE Candidate send it to the remote peer
	if err := p.agent.OnCandidate(p.onLocalCandidate); err != nil {
		return fmt.Errorf("failed to setup on candidate handler: %w", err)
	}

	// When selected candidate pair changes
	if err := p.agent.OnSelectedCandidatePairChange(p.onSelectedCandidatePairChange); err != nil {
		return fmt.Errorf("failed to setup on selected candidate pair handler: %w", err)
	}

	// When ICE Connection state has change print to stdout
	if err := p.agent.OnConnectionStateChange(p.onConnectionStateChange); err != nil {
		return fmt.Errorf("failed to setup on connection state handler: %w", err)
	}

	if _, ok := p.connectionState.SetIf(ConnectionStateIdle, ConnectionStateCreating); !ok {
		return errSwitchToIdle
	}

	// Send peer credentials as long as we remain in ConnectionStateIdle
	go p.sendCredentialsWhileIdleWithBackoff(true)

	return nil
}

func (p *Peer) sendCredentialsWhileIdleWithBackoff(need bool) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxInterval = 1 * time.Minute

	if err := backoff.RetryNotify(
		func() error {
			if p.connectionState.Load() != ConnectionStateIdle {
				// We are not idling any more.
				// No need to send credentials
				return nil
			}

			if err := p.sendCredentials(need); err != nil {
				if errors.Is(err, signaling.ErrClosed) {
					// Do not retry when the signaling backend has been closed
					return nil
				}

				return err
			}

			return errStillIdle
		}, bo,
		func(err error, d time.Duration) {
			if errors.Is(err, errStillIdle) {
				p.logger.Debug("Sending peer credentials while waiting for remote peer",
					zap.Error(err),
					zap.Duration("after", d))
			} else if sts := status.Code(err); sts != codes.Canceled {
				p.logger.Error("Failed to send peer credentials",
					zap.Error(err),
					zap.Duration("after", d))
			}
		},
	); err != nil {
		p.logger.Error("Failed to send credentials", zap.Error(err))
	}
}

// isSessionRestart checks if a received offer should restart the
// ICE session by comparing ufrag & pwd with previously used values.
func (p *Peer) isSessionRestart(c *epdiscproto.Credentials) bool {
	r := p.remoteCredentials

	return (r != nil) &&
		(r.Ufrag != "" && r.Pwd != "") &&
		(c.Ufrag != "" && c.Pwd != "") &&
		(r.Ufrag != c.Ufrag || r.Pwd != c.Pwd)
}

func (p *Peer) connect(ufrag, pwd string) {
	var connect func(context.Context, string, string) (*ice.Conn, error)
	if p.IsControlling() {
		p.logger.Debug("Dialing...")
		connect = p.agent.Dial
	} else {
		p.logger.Debug("Accepting...")
		connect = p.agent.Accept
	}

	if conn, err := connect(context.TODO(), ufrag, pwd); err == nil {
		p.newConnection <- conn
	} else {
		p.logger.Error("Failed to connect", zap.Error(err))
	}
}

func (p *Peer) updateProxy(cp *ice.CandidatePair, conn *ice.Conn) error {
	var err error
	var oldProxy, newProxy Proxy
	var newEndpoint *net.UDPAddr

	bind := p.Interface.Bind()

	// Create new proxy
	switch {
	case p.Interface.IsUserspace():
		p.logger.Debug("Forwarding via in-process bind")
		newProxy, newEndpoint, err = NewBindProxy(bind, cp, conn, p.logger)
	case p.Interface.nat != nil && CandidatePairCanBeNATted(cp):
		p.logger.Debug("Forwarding via kernel NAT")
		newProxy, newEndpoint, err = NewKernelNATProxy(cp, p.Interface.nat, p.Interface.ListenPort, p.logger)
	default:
		p.logger.Debug("Forwarding via kernel connection")
		newProxy, newEndpoint, err = NewKernelConnProxy(bind, cp, conn, p.Interface.ListenPort, p.logger)
	}
	if err != nil {
		return fmt.Errorf("failed to setup proxy: %w", err)
	}

	// Check if we need to update the bind
	updateBind := false

	// Close old proxy
	if oldProxy = p.proxy; oldProxy != nil {
		if err := oldProxy.Close(); err != nil {
			return fmt.Errorf("failed to close old proxy: %w", err)
		}

		if _, ok := oldProxy.(wg.BindConn); ok {
			updateBind = true
		}
	}

	if _, ok := newProxy.(wg.BindConn); ok {
		updateBind = true
	}

	p.endpoint = newEndpoint
	p.proxy = newProxy

	if updateBind {
		if err := p.Interface.Device.BindUpdate(); err != nil {
			return fmt.Errorf("failed to update bind: %w", err)
		}
	}

	if err := p.SetEndpoint(p.endpoint); err != nil {
		return fmt.Errorf("failed to update endpoint: %w", err)
	}

	return nil
}
