package hsync

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (i *Interface) OnPeerAdded(p *core.Peer) {
	if err := i.Sync(); err != nil {
		i.logger.Error("Failed to update hosts file", zap.Error(err))
	}

	p.OnModified(i)
}

func (i *Interface) OnPeerRemoved(p *core.Peer) {
	if err := i.Sync(); err != nil {
		i.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (i *Interface) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	// Only update if the name has changed
	if m.Is(core.PeerModifiedName) {
		if err := i.Sync(); err != nil {
			i.logger.Error("Failed to update hosts file", zap.Error(err))
		}
	}
}
