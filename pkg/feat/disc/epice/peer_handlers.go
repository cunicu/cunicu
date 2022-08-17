package epice

import (
	"errors"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	icex "riasc.eu/wice/pkg/ice"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

// onConnectionStateChange is a callback which gets called by the ICE agent
// whenever the state of the ICE connection has changed
func (p *Peer) onConnectionStateChange(cs ice.ConnectionState) {
	var err error

	csx := icex.ConnectionState(cs)
	prevConnectionState := p.setConnectionState(csx)

	if cs == ice.ConnectionStateFailed || cs == ice.ConnectionStateDisconnected {
		// TODO: Add some random delay?

		if err := p.Restart(); err != nil {
			p.logger.Error("Failed to restart ICE session", zap.Error(err))
		}
	} else if cs == ice.ConnectionStateClosed && prevConnectionState != icex.ConnectionStateClosing {
		if p.agent, err = p.newAgent(); err != nil {
			p.logger.Error("Failed to create agent", zap.Error(err))
			return
		}

		if err := p.sendCredentials(true); err != nil {
			p.logger.Error("Failed to send peer credentials", zap.Error(err))
			return
		}
	}
}

// onCandidate is a callback which gets called for each discovered local ICE candidate
func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
	} else {
		p.logger.Info("Found new local candidate", zap.Any("candidate", c))

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
func (p *Peer) onRemoteCredentials(c *pb.Credentials) {
	logger := p.logger.With(zap.Any("creds", c))
	logger.Info("Received remote credentials")

	if p.isSessionRestart(c) {
		if err := p.Restart(); err != nil {
			p.logger.Error("Failed to restart ICE session", zap.Error(err))
		}
	} else {
		if p.ConnectionState == icex.ConnectionStateIdle {
			p.setConnectionState(ice.ConnectionStateNew)

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

		if c.NeedCreds {
			if err := p.sendCredentials(false); err != nil {
				p.logger.Error("Failed to send credentials", zap.Error(err))
				return
			}
		}
	}
}

// onRemoteCandidate is a handler called for each received candidate via the signaling channel
func (p *Peer) onRemoteCandidate(c *pb.Candidate) {
	logger := p.logger.With(zap.Any("candidate", c))
	logger.Info("Received remote candidate")

	if err := p.addRemoteCandidate(c); err != nil {
		p.logger.Error("Failed to add candidates", zap.Error(err))
		return
	}

	if p.ConnectionState == ice.ConnectionStateNew {
		p.setConnectionState(icex.ConnectionStateConnecting)

		ufrag, pwd, err := p.agent.GetRemoteUserCredentials()
		if err != nil {
			p.logger.Error("Failed to get remote credentials", zap.Error(err))
			return
		}

		if err := p.connect(ufrag, pwd); err != nil && !errors.Is(err, ice.ErrClosed) {
			p.logger.Error("Failed to connect", zap.Error(err))
			return
		}
	}
}

func (p *Peer) onSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if msg.Credentials != nil {
		p.onRemoteCredentials(msg.Credentials)
	}

	if msg.Candidate != nil {
		p.onRemoteCandidate(msg.Candidate)
	}
}

// OnSignalingMessage is invoked for every message received via the signaling backend
func (p *Peer) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if p.agent == nil {
		p.logger.Warn("Ignoring message as agent has not been created yet")
	} else {
		go p.onSignalingMessage(kp, msg)
	}
}
