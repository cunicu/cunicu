package p2p

import "github.com/libp2p/go-libp2p-core/peer"

const (
	mdnsServiceName = "wice/0.1"
)

type mDNSNotifee struct {
	backend *Backend
}

func (m *mDNSNotifee) HandlePeerFound(ai peer.AddrInfo) {
	m.backend.handleMDNSPeer(ai)
}
