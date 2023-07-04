// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"github.com/pion/ice/v2"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/wg"
)

// onConnectionStateChange is a callback which gets called by the ICE agent
// whenever the state of the ICE connection has changed
// It is started as goroutine from pion/ice.Agent.
func (p *Peer) onConnectionStateChange(ics ice.ConnectionState) {
	cs := epdiscproto.NewConnectionState(ics)

	switch cs {
	case ConnectionStateFailed, ConnectionStateDisconnected:
		if cs, ok := p.SetStateIf(daemon.PeerStateFailed, daemon.PeerStateConnected); !ok {
			p.logger.Error("Invalid state transition",
				zap.Any("current_state", cs),
				zap.Any("new_state", daemon.PeerStateFailed))
		}

		if err := p.Restart(); err != nil {
			p.logger.Error("Failed to restart ICE session", zap.Error(err))
		}

	case ConnectionStateClosed:
		if _, ok := p.connectionState.SetIf(ConnectionStateClosed, ConnectionStateClosing); ok {
			// Peer is now closed
			// TODO: Stop run() goroutine?
			break
		} else if _, ok := p.connectionState.SetIf(ConnectionStateClosed, ConnectionStateRestarting); ok {
			go p.createAgentWithBackoff()
		}

	case ConnectionStateConnected:
		if ps, ok := p.connectionState.SetIf(ConnectionStateConnected, ConnectionStateConnecting); !ok {
			p.logger.Error("Invalid state transition",
				zap.Any("current_state", ps),
				zap.Any("new_state", ConnectionStateConnected))
		}

		cp, err := p.agent.GetSelectedCandidatePair()
		if err != nil {
			p.logger.Error("Failed to get selected candidate pair", zap.Error(err))
			break
		}

		if err := p.updateProxy(cp, <-p.newConnection); err != nil {
			p.logger.Error("Failed to update proxy", zap.Error(err))
			break
		}

		// Signal to daemon that we are now connected
		if _, ok := p.SetStateIf(daemon.PeerStateConnected, daemon.PeerStateConnecting); !ok {
			p.logger.Error("Invalid state transition",
				zap.Any("current_state", cs),
				zap.Any("new_state", daemon.PeerStateConnected))
		}

	default:
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

// onRemoteCredentials is a handler called for each received pair of remote Ufrag/Pwd via the signaling channel
func (p *Peer) onRemoteCredentials(creds *epdiscproto.Credentials) {
	logger := p.logger.With(zap.Reflect("creds", creds))
	logger.Debug("Received remote credentials")

	if p.isSessionRestart(creds) {
		if err := p.Restart(); err != nil {
			p.logger.Error("Failed to restart ICE session", zap.Error(err))
		}
	} else {
		if _, ok := p.connectionState.SetIf(ConnectionStateGathering, ConnectionStateIdle); !ok {
			p.logger.Debug("Ignoring duplicated credentials")
			return
		}

		p.SetStateIf(daemon.PeerStateConnecting, daemon.PeerStateClosed, daemon.PeerStateFailed, daemon.PeerStateNew)

		p.remoteCredentials = creds

		// Return our own credentials if requested
		if creds.NeedCreds {
			if err := p.sendCredentials(false); err != nil {
				p.logger.Error("Failed to send credentials", zap.Error(err))
				return
			}
		}

		// Start gathering candidates
		if err := p.agent.GatherCandidates(); err != nil {
			p.logger.Error("failed to gather candidates", zap.Error(err))
			return
		}
	}
}

// onRemoteCandidate is a handler called for each received candidate via the signaling channel
func (p *Peer) onRemoteCandidate(c *epdiscproto.Candidate) {
	logger := p.logger.With(zap.Reflect("candidate", c))

	ic, err := c.ICECandidate()
	if err != nil {
		logger.Error("Failed to remote candidate", zap.Error(err))
		return
	}

	if err := p.agent.AddRemoteCandidate(ic); err != nil {
		logger.Error("Failed to add remote candidate", zap.Error(err))
		return
	}

	logger.Debug("Added remote candidate to agent")

	// Connect if this has been the first remote candidate
	if _, ok := p.connectionState.SetIf(ConnectionStateConnecting, ConnectionStateGatheringRemote); ok {
		go p.connect(p.remoteCredentials.Ufrag, p.remoteCredentials.Pwd)
	} else {
		// Continue gathering until we found the first local candidate
		p.connectionState.SetIf(ConnectionStateGatheringLocal, ConnectionStateGathering)
	}
}

// onLocalCandidate is a callback which gets called for each discovered local ICE candidate
func (p *Peer) onLocalCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
		return
	}

	logger := p.logger.With(zap.Reflect("candidate", c))
	logger.Debug("Added local candidate to agent")

	if err := p.sendCandidate(c); err != nil {
		logger.Error("Failed to send candidate", zap.Error(err))
	}

	if _, ok := p.connectionState.SetIf(ConnectionStateConnecting, ConnectionStateGatheringLocal); ok {
		go p.connect(p.remoteCredentials.Ufrag, p.remoteCredentials.Pwd)
	} else {
		// Continue waiting until we received the first remote candidate
		p.connectionState.SetIf(ConnectionStateGatheringRemote, ConnectionStateGathering)
	}
}

// onSignalingMessage is invoked for every message received via the signaling backend
func (p *Peer) onSignalingMessage(msg *signaling.Message) {
	if msg.Credentials != nil {
		p.onRemoteCredentials(msg.Credentials)
	}

	if msg.Candidate != nil {
		p.onRemoteCandidate(msg.Candidate)
	}
}

// OnSignalingMessage is invoked for every message received via the signaling backend
func (p *Peer) OnSignalingMessage(_ *crypto.PublicKeyPair, msg *signaling.Message) {
	p.signalingMessages <- msg
}

func (p *Peer) OnBindOpen(b *wg.Bind, _ uint16) {
	if conn, ok := p.proxy.(wg.BindConn); ok {
		b.Conns = append(b.Conns, conn)
	}
}
