package pske

import (
	"github.com/stv0g/cunicu/pkg/core"
	"go.uber.org/zap"
)

func (i *Interface) OnPeerAdded(cp *core.Peer) {
	if cp.PresharedKey().IsSet() {
		i.logger.Debug("Ignoring peer as it already has a PSK configured", zap.Any("peer", cp))
		return
	}

	i.Peers[cp] = i.NewPeer(cp, i)
}

func (i *Interface) OnPeerRemoved(p *core.Peer) {}
