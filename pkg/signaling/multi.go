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

func (m *MultiBackend) Publish(kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	for _, b := range m.backends {
		if err := b.Publish(kp, msg); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultiBackend) Subscribe(kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	chans := []chan *pb.SignalingMessage{}

	for _, b := range m.backends {
		if ch, err := b.Subscribe(kp); err != nil {
			return nil, err
		} else {
			chans = append(chans, ch)
		}
	}

	return pumpMessages(chans), nil
}

func (m *MultiBackend) Close() error {
	for _, b := range m.backends {
		if err := b.Close(); err != nil {
			return err
		}
	}

	return nil
}

// pumpMessages reads offers from the secondary backends and pushes them into a common channel
func pumpMessages(chans []chan *pb.SignalingMessage) chan *pb.SignalingMessage {
	nch := make(chan *pb.SignalingMessage)

	for _, ch := range chans {
		go func(ch chan *pb.SignalingMessage) {
			for m := range ch {
				nch <- m
			}
		}(ch)
	}

	return nch
}
