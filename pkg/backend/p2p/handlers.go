package p2p

import (
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

func (b *Backend) handleStream(stream network.Stream) {
	var err error

	peerID := stream.Conn().RemotePeer()
	peer := b.peers.GetByPeerId(peerID)
	if peer == nil {
		peer, err = NewPeer(b, peerID)
		if err != nil {
			b.Logger.WithError(err).Fatal("Failed to create peer")
			return
		}
	}

	peer.HandleStream(stream)
}

func (b *Backend) handlePeers(peerChan <-chan peer.AddrInfo) {
	var err error

	for peer := range peerChan {
		if peer.ID == b.host.ID() {
			continue
		}

		p := b.peers.GetByPeerId(peer.ID)
		if p == nil {
			p, err = NewPeer(b, peer.ID)
			if err != nil {
				b.Logger.WithError(err).Error("Failed to create peer")
			}

			b.peers = append(b.peers, p)
		}
	}
}
