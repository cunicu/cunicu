//go:build linux

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
