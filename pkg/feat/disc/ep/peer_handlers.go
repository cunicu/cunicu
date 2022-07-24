package ep

import (
	"fmt"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

// onCandidate is a callback which gets called for each discovered local ICE candidate
func (p *Peer) onCandidate(c ice.Candidate) {
	if c == nil {
		p.logger.Info("Candidate gathering completed")
	} else {
		p.logger.Info("Found new local candidate", zap.Any("candidate", c))

		p.description.Candidates = append(p.description.Candidates, pb.NewCandidate(c))

		if err := p.sendDescription(); err != nil {
			p.logger.Error("Failed to send description", zap.Error(err))
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

// onOffer is a handler called for each received offer via the signaling channel
func (p *Peer) onDescription(sd *pb.SessionDescription) error {
	logger := p.logger.With(zap.Any("description", sd))
	logger.Info("Received session description")

	if p.isSessionRestart(sd) {
		if err := p.restart(); err != nil {
			return fmt.Errorf("failed to restart: %w", err)
		}
	}

	if err := p.addCandidates(sd); err != nil {
		return fmt.Errorf("failed to add candidates: %w", err)
	}

	if p.ConnectionState == ice.ConnectionStateNew {
		go func() {
			if err := p.connect(sd.Ufrag, sd.Pwd); err != nil {
				p.logger.Error("Failed to connect", zap.Error(err))
			}
		}()
	}

	return nil
}

// onMessage is called for each received message via the signaling channel
func (p *Peer) onMessage(msg *signaling.Message) error {
	switch {
	case msg.Session != nil:
		if err := p.onDescription(msg.Session); err != nil {
			return err
		}
	}

	return nil
}
