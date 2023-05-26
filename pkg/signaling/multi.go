// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package signaling

import (
	"context"
	"net/url"

	"github.com/stv0g/cunicu/pkg/crypto"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
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

func (mb *MultiBackend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error) {
	for _, b := range mb.Backends {
		if _, err := b.Subscribe(ctx, kp, h); err != nil {
			return false, err
		}
	}

	return false, nil
}

func (mb *MultiBackend) Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error) {
	for _, b := range mb.Backends {
		if _, err := b.Unsubscribe(ctx, kp, h); err != nil {
			return false, err
		}
	}

	return false, nil
}

func (mb *MultiBackend) Close() error {
	for _, b := range mb.Backends {
		if err := b.Close(); err != nil {
			return err
		}
	}

	return nil
}
