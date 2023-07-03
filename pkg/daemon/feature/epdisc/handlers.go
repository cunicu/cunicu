// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"net"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/wg"
)

func (i *Interface) OnInterfaceModified(_ *daemon.Interface, _ *wg.Interface, m daemon.InterfaceModifier) {
	if m.Is(daemon.InterfaceModifiedListenPort) {
		if err := i.updateNATRules(); err != nil {
			i.logger.Error("Failed to update NAT rules", zap.Error(err))
		}
	}
}

func (i *Interface) OnPeerAdded(cp *daemon.Peer) {
	p, err := NewPeer(cp, i)
	if err != nil {
		i.logger.Error("Failed to initialize ICE peer", zap.Error(err))
		return
	}

	i.Peers[cp] = p
}

func (i *Interface) OnPeerRemoved(cp *daemon.Peer) {
	p, ok := i.Peers[cp]
	if !ok {
		return
	}

	if err := p.Close(); err != nil {
		i.logger.Error("Failed to de-initialize ICE peer", zap.Error(err))
	}

	delete(i.Peers, cp)
}

func (i *Interface) OnPeerModified(cp *daemon.Peer, _ *wgtypes.Peer, m daemon.PeerModifier, _, _ []net.IPNet) {
	p := i.Peers[cp]

	if m.Is(daemon.PeerModifiedEndpoint) {
		// Check if change was external
		epNew := p.Endpoint
		epExpected := p.endpoint

		if (epExpected != nil && epNew != nil) && (!epNew.IP.Equal(epExpected.IP) || epNew.Port != epExpected.Port) {
			i.logger.Warn("Endpoint address has been changed externally. This is breaks the connection and is most likely not desired.")
		}
	}
}

func (i *Interface) OnBindOpen(b *wg.Bind, _ uint16) {
	for _, muxConn := range i.muxConns {
		b.AddConn(muxConn)
	}
}
