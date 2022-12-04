package epdisc

import (
	"errors"

	"github.com/pion/ice/v2"
	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"go.uber.org/zap"
)

// onConnectionStateChange is a callback which gets called by the ICE agent
// whenever the state of the ICE connection has changed
func (p *Peer) onConnectionStateChange(newState icex.ConnectionState) {
	if p.ConnectionState() == icex.ConnectionStateClosing {
		p.logger.Debug("Ignoring state transition as we are closing the session")
		return
	}

	p.setConnectionState(newState)

	if newState == ice.ConnectionStateFailed || newState == ice.ConnectionStateDisconnected {
		if err := p.Restart(); err != nil {
			p.logger.Error("Failed to restart ICE session", zap.Error(err))
		}
	} else if newState == ice.ConnectionStateClosed {
		go func() {
			if err := p.createAgentWithBackoff(); err != nil {
				p.logger.Error("Failed to connect", zap.Error(err))
			}
		}()
	}
}

// onCandidate is a callback which gets called for each discovered local ICE candidate
func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
	} else {
		p.logger.Debug("Found new local candidate", zap.Any("candidate", c))

		if err := p.sendCandidate(c); err != nil {
			p.logger.Error("Failed to send candidate", zap.Error(err))
		}
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
func (p *Peer) onRemoteCredentials(c *epdiscproto.Credentials) {
	logger := p.logger.With(zap.Any("creds", c))
	logger.Info("Received remote credentials")

	if p.isSessionRestart(c) {
		if err := p.Restart(); err != nil {
			p.logger.Error("Failed to restart ICE session", zap.Error(err))
		}
	} else {
		if c.NeedCreds {
			if err := p.sendCredentials(false); err != nil {
				p.logger.Error("Failed to send credentials", zap.Error(err))
				return
			}
		}

		if p.setConnectionStateIf(icex.ConnectionStateIdle, ice.ConnectionStateNew) {
			if err := p.agent.SetRemoteCredentials(c.Ufrag, c.Pwd); err != nil {
				p.logger.Error("Failed to set remote credentials", zap.Error(err))
				return
			}

			if err := p.agent.GatherCandidates(); err != nil {
				p.logger.Error("Failed to gather candidates", zap.Error(err))
				return
			}
			p.logger.Info("Started gathering local ICE candidates")
		}
	}
}

// onRemoteCandidate is a handler called for each received candidate via the signaling channel
func (p *Peer) onRemoteCandidate(c *epdiscproto.Candidate) {
	logger := p.logger.With(zap.Any("candidate", c))
	logger.Debug("Received remote candidate")

	if err := p.addRemoteCandidate(c); err != nil {
		p.logger.Error("Failed to add candidates", zap.Error(err))
		return
	}

	if p.setConnectionStateIf(ice.ConnectionStateNew, icex.ConnectionStateConnecting) {
		ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
		if err != nil {
			p.logger.Error("Failed to get remote credentials", zap.Error(err))
			return
		}

		go func() {
			if err := p.connect(ufrag, pwd); err != nil && !errors.Is(err, ice.ErrClosed) {
				p.logger.Error("Failed to connect", zap.Error(err))
				return
			}
		}()
	}
}

func (p *Peer) onSignalingMessage(msg *signaling.Message) {
	if msg.Credentials != nil {
		p.onRemoteCredentials(msg.Credentials)
	}

	if msg.Candidate != nil {
		p.onRemoteCandidate(msg.Candidate)
	}
}

// OnSignalingMessage is invoked for every message received via the signaling backend
func (p *Peer) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	p.signalingMessages <- msg
}
