// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type WireGuardPeerOption interface {
	Apply(i *WireGuardPeer)
}

type WireGuardPeer struct {
	wgtypes.PeerConfig
}

func (p *WireGuardPeer) Apply(i *WireGuardInterface) {
	i.Peers = append(i.Peers, p.PeerConfig)
}
