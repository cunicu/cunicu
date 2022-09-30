package hsync

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (hs *Interface) OnPeerAdded(p *core.Peer) {
	if err := hs.Sync(); err != nil {
		hs.logger.Error("Failed to update hosts file", zap.Error(err))
	}

	p.OnModified(hs)
}

func (hs *Interface) OnPeerRemoved(p *core.Peer) {
	if err := hs.Sync(); err != nil {
		hs.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (hs *Interface) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	// Only update if the name has changed
	if m.Is(core.PeerModifiedName) {
		if err := hs.Sync(); err != nil {
			hs.logger.Error("Failed to update hosts file", zap.Error(err))
		}
	}
}
