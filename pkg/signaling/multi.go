package signaling

import (
	"context"
	"errors"
	"net/url"

	"riasc.eu/wice/pkg/crypto"

	signalingproto "riasc.eu/wice/pkg/proto/signaling"
)

type MultiBackend struct {
	Backends []Backend
}

func NewMultiBackend(uris []*url.URL, cfg *BackendConfig) (*MultiBackend, error) {
	mb := &MultiBackend{
		Backends: []Backend{},
	}

	for _, u := range uris {
		cfg.URI = u

		if b, err := NewBackend(cfg); err == nil {
			mb.Backends = append(mb.Backends, b)
		} else {
			return nil, err
		}
	}

	return mb, nil
}

func (mb *MultiBackend) Type() signalingproto.BackendType {
	return signalingproto.BackendType_MULTI
}

func (mb *MultiBackend) ByType(t signalingproto.BackendType) Backend {
	for _, b := range mb.Backends {
		if b.Type() == t {
			return b
		}
	}

	return nil
}

func (mb *MultiBackend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *Message) error {
	for _, b := range mb.Backends {
		if err := b.Publish(ctx, kp, msg); err != nil {
			return err
		}
	}

	return nil
}

func (mb *MultiBackend) SubscribeAll(ctx context.Context, kp *crypto.Key, h MessageHandler) error {
	return errors.New("not implemented yet") // TODO
}

func (mb *MultiBackend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) error {
	for _, b := range mb.Backends {
		if err := b.Subscribe(ctx, kp, h); err != nil {
			return err
		}
	}

	return nil
}

func (mb *MultiBackend) Close() error {
	for _, b := range mb.Backends {
		if err := b.Close(); err != nil {
			return err
		}
	}

	return nil
}
