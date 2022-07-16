package core

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// Peer added/removed

type PeerHandlerList []PeerHandler
type PeerHandler interface {
	OnPeerAdded(p *Peer)
	OnPeerRemoved(p *Peer)
}

func (hl *PeerHandlerList) Register(h PeerHandler) {
	*hl = append(*hl, h)
}

func (hl *PeerHandlerList) InvokeAdded(p *Peer) {
	for _, h := range *hl {
		h.OnPeerAdded(p)
	}
}

func (hl *PeerHandlerList) InvokeRemoved(p *Peer) {
	for _, h := range *hl {
		h.OnPeerAdded(p)
	}
}

// Peer modified

type PeerModifiedHandlerList []PeerModifiedHandler
type PeerModifiedHandler interface {
	OnPeerModified(p *Peer, old *wgtypes.Peer, m PeerModifier, ipsAdded, ipsRemoved []net.IPNet)
}

func (hl *PeerModifiedHandlerList) Register(h PeerModifiedHandler) {
	*hl = append(*hl, h)
}

func (hl *PeerModifiedHandlerList) Invoke(p *Peer, old *wgtypes.Peer, mod PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	for _, h := range *hl {
		h.OnPeerModified(p, old, mod, ipsAdded, ipsRemoved)
	}
}
