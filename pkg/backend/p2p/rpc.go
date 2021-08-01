package p2p

import (
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/crypto"
)

type CandidateService struct {
	Peer *Peer
}

func NewCandidateService(p *Peer) *CandidateService {
	return &CandidateService{
		Peer: p,
	}
}

func (svc *CandidateService) Publish(om backend.OfferMap, result *bool) error {

	for pk, o := range om {
		kp := crypto.PublicKeyPair{
			Ours:   svc.Peer.PublicKey,
			Theirs: pk,
		}

		ch, ok := svc.Peer.Backend.Offers[kp]
		if ok {
			ch <- o
		}
	}

	return nil
}

func (svc *CandidateService) Remove(pk crypto.Key, result *bool) error {
	return nil
}

type PeerService struct {
	Peer *Peer
}

func NewPeerService(p *Peer) *PeerService {
	return &PeerService{
		Peer: p,
	}
}

func (svc *PeerService) Add(p backend.Peer, result *bool) error {
	return nil
}

func (svc *PeerService) Remove(p backend.Peer, result *bool) error {
	return nil
}
