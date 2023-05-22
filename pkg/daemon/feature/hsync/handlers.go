package hsync

import (
	"net"

	"github.com/stv0g/cunicu/pkg/daemon"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (i *Interface) OnPeerAdded(p *daemon.Peer) {
	if err := i.Sync(); err != nil {
		i.logger.Error("Failed to update hosts file", zap.Error(err))
	}

	p.AddModifiedHandler(i)
}

func (i *Interface) OnPeerRemoved(p *daemon.Peer) {
	if err := i.Sync(); err != nil {
		i.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (i *Interface) OnPeerModified(p *daemon.Peer, old *wgtypes.Peer, m daemon.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	// Only update if the name has changed
	if m.Is(daemon.PeerModifiedName) {
		if err := i.Sync(); err != nil {
			i.logger.Error("Failed to update hosts file", zap.Error(err))
		}
	}
}
