package mtudisc

import (
	"github.com/stv0g/cunicu/pkg/core"
	"go.uber.org/zap"
)

func (i *Interface) OnPeerAdded(cp *core.Peer) {
	i.peers[cp] = i.NewPeer(cp)
}

func (i *Interface) OnPeerRemoved(cp *core.Peer) {
	p := i.peers[cp]

	if err := p.Close(); err != nil {
		i.logger.Error("Failed to close peer", zap.Error(err))
	}

	delete(i.peers, cp)

	if err := i.UpdateMTU(); err != nil {
		i.logger.Error("Failed to update MTU", zap.Error(err))
	}
}
