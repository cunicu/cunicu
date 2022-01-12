package signaling

import (
	"net/url"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type MultiBackend struct {
	backends []Backend
}

func NewMultiBackend(uris []*url.URL, events chan *pb.Event) (Backend, error) {
	mb := &MultiBackend{
		backends: []Backend{},
	}

	for _, u := range uris {
		b, err := NewBackend(u, events)
		if err != nil {
			return nil, err
		}

		mb.backends = append(mb.backends, b)
	}

	return mb, nil
}

func (m *MultiBackend) PublishOffer(kp crypto.PublicKeyPair, offer *pb.Offer) error {
	for _, b := range m.backends {
		if err := b.PublishOffer(kp, offer); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultiBackend) SubscribeOffer(kp crypto.PublicKeyPair) (chan *pb.Offer, error) {
	chans := []chan *pb.Offer{}

	for _, b := range m.backends {
		if ch, err := b.SubscribeOffer(kp); err != nil {
			return nil, err
		} else {
			chans = append(chans, ch)
		}
	}

	return pumpOffers(chans), nil
}

func (m *MultiBackend) Tick() {

}

func (m *MultiBackend) Close() error {
	for _, b := range m.backends {
		if err := b.Close(); err != nil {
			return err
		}
	}

	return nil
}

// pumpOffers reads offers from the secondary backends and pushes them into a common channel
func pumpOffers(chans []chan *pb.Offer) chan *pb.Offer {
	nch := make(chan *pb.Offer)

	for _, ch := range chans {
		go func(ch chan *pb.Offer) {
			for o := range ch {
				nch <- o
			}
		}(ch)
	}

	return nch
}
