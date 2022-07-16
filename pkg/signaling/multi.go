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

func NewMultiBackend(uris []*url.URL, cfg *BackendConfig) (Backend, error) {
	mb := &MultiBackend{
		backends: []Backend{},
	}

	for _, u := range uris {
		cfg.URI = u

		if b, err := NewBackend(cfg); err == nil {
			mb.backends = append(mb.backends, b)
		} else {
			return nil, err
		}
	}

	return mb, nil
}

func (mb *MultiBackend) OnReady(h BackendReadyHandler) {
	for _, b := range mb.backends {
		b.OnReady(h)
	}
}

func (mb *MultiBackend) Type() pb.BackendReadyEvent_Type {
	return pb.BackendReadyEvent_MULTI
}

func (mb *MultiBackend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	for _, b := range mb.backends {
		if err := b.Publish(ctx, kp, msg); err != nil {
			return err
		}
	}

	return nil
}

func (mb *MultiBackend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	chans := []chan *pb.SignalingMessage{}

	for _, b := range mb.backends {
		ch, err := b.Subscribe(ctx, kp)
		if err != nil {
			return nil, err
		}

		chans = append(chans, ch)
	}

	return util.FanIn(chans...), nil
}

func (mb *MultiBackend) Close() error {
	for _, b := range mb.backends {
		if err := b.Close(); err != nil {
			return err
		}
	}

	return nil
}
