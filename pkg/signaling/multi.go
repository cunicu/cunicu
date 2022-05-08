package signaling

import (
	"context"
	"net/url"

	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type MultiBackend struct {
	backends []Backend
}

func NewMultiBackend(uris []*url.URL, cfg *BackendConfig, events chan *pb.Event) (Backend, error) {
	mb := &MultiBackend{
		backends: []Backend{},
	}

	for _, u := range uris {
		cfg.URI = u

		if b, err := NewBackend(cfg, events); err == nil {
			mb.backends = append(mb.backends, b)
		} else {
			return nil, err
		}
	}

	return mb, nil
}

func (m *MultiBackend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	for _, b := range m.backends {
		if err := b.Publish(ctx, kp, msg); err != nil {
			return err
		}
	}

	return nil
}

func (m *MultiBackend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	chans := []chan *pb.SignalingMessage{}

	for _, b := range m.backends {
		ch, err := b.Subscribe(ctx, kp)
		if err != nil {
			return nil, err
		}

		chans = append(chans, ch)
	}

	return util.FanIn(chans...), nil
}

func (m *MultiBackend) Close() error {
	for _, b := range m.backends {
		if err := b.Close(); err != nil {
			return err
		}
	}

	return nil
}
