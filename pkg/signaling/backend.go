// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package signaling

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/log"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
)

var (
	Backends = map[BackendType]*BackendPlugin{} //nolint:gochecknoglobals

	ErrInvalidBackend = errors.New("unknown backend type")
	ErrClosed         = errors.New("backend is closed")
)

type BackendType string // URL schemes

type BackendFactory func(*BackendConfig, *log.Logger) (Backend, error)

type BackendPlugin struct {
	New         BackendFactory
	Description string
}

type BackendConfig struct {
	URI *url.URL

	OnReady []BackendReadyHandler
}

type BackendReadyHandler interface {
	OnSignalingBackendReady(b Backend)
}

type Backend interface {
	io.Closer

	// Publish a signaling message to a specific peer
	Publish(ctx context.Context, kp *crypto.KeyPair, msg *Message) error

	// Subscribe to messages send by a specific peer
	Subscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error)

	// Unsubscribe from messages send by a specific peer
	Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h MessageHandler) (bool, error)

	// Returns the backends type identifier
	Type() signalingproto.BackendType
}

func NewBackend(cfg *BackendConfig) (Backend, error) {
	typs := strings.SplitN(cfg.URI.Scheme, "+", 2)
	typ := BackendType(typs[0])

	p, ok := Backends[typ]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrInvalidBackend, typ)
	}

	if len(typs) > 1 {
		cfg.URI.Scheme = typs[1]
	}

	loggerName := fmt.Sprintf("backend.%s", typ)
	logger := log.Global.Named(loggerName).With(zap.Any("backend", cfg.URI))

	be, err := p.New(cfg, logger)
	if err != nil {
		return nil, err
	}

	return be, nil
}
