// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"net"

	"golang.org/x/exp/slices"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type PeerModifiedHandler interface {
	OnPeerModified(p *Peer, old *wgtypes.Peer, m PeerModifier, ipsAdded, ipsRemoved []net.IPNet)
}

// AddModifiedHandler registers a new handler which is called whenever the peer has been modified
func (p *Peer) AddModifiedHandler(h PeerModifiedHandler) {
	if !slices.Contains(p.onModified, h) {
		p.onModified = append(p.onModified, h)
	}
}

func (p *Peer) RemoveModifiedHandler(h PeerModifiedHandler) {
	if idx := slices.Index(p.onModified, h); idx > -1 {
		p.onModified = slices.Delete(p.onModified, idx, idx+1)
	}
}
