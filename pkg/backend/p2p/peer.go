package p2p

import (
	"fmt"
	"net/rpc"

	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/crypto"
)

type Peer struct {
	ID     peer.ID
	Stream network.Stream

	Backend   *Backend
	PublicKey crypto.Key

	Client *rpc.Client

	logger *log.Entry
}

type PeerList []*Peer

func (pl *PeerList) GetByPeerId(id peer.ID) *Peer {
	for _, peer := range *pl {
		if peer.ID == id {
			return peer
		}
	}

	return nil
}

func (pl *PeerList) GetByPublicKey(pk crypto.Key) *Peer {
	for _, peer := range *pl {
		if peer.PublicKey == pk {
			return peer
		}
	}

	return nil
}

func NewPeer(b *Backend, id peer.ID) (*Peer, error) {
	var err error

	p := Peer{
		ID:      id,
		Backend: b,
		logger:  b.Logger.WithField("peer", id),
	}

	p.logger.Debug("Connecting to peer")

	p.Stream, err = b.host.NewStream(b.context, p.ID, ProtocolID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer: %w", err)
	}

	p.logger.Info("Connected to peer")

	p.Client = rpc.NewClient(p.Stream)

	return &p, nil
}

func (p *Peer) Close() error {
	if p.Stream != nil {
		return p.Stream.Close()
	}

	return nil
}

func (p *Peer) HandleStream(stream network.Stream) {
	server := rpc.NewServer()

	server.RegisterName("candidates", NewCandidateService(p))
	server.RegisterName("peers", NewPeerService(p))

	go server.ServeConn(stream)
}
