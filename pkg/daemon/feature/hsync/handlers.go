// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package hsync

import (
	"net"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/daemon"
)

func (i *Interface) OnPeerAdded(p *daemon.Peer) {
	if err := i.Sync(); err != nil {
		i.logger.Error("Failed to update hosts file", zap.Error(err))
	}

	p.AddModifiedHandler(i)
}

func (i *Interface) OnPeerRemoved(_ *daemon.Peer) {
	if err := i.Sync(); err != nil {
		i.logger.Error("Failed to update hosts file", zap.Error(err))
	}
}

func (i *Interface) OnPeerModified(_ *daemon.Peer, _ *wgtypes.Peer, m daemon.PeerModifier, _, _ []net.IPNet) {
	// Only update if the name has changed
	if m.Is(daemon.PeerModifiedName) {
		if err := i.Sync(); err != nil {
			i.logger.Error("Failed to update hosts file", zap.Error(err))
		}
	}
}
