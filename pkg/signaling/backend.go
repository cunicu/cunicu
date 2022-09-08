package signaling

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"

	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
)

var (
	Backends = map[BackendType]*BackendPlugin{}
)

type BackendType string // URL schemes

type BackendFactory func(*BackendConfig, *zap.Logger) (Backend, error)

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
		return nil, fmt.Errorf("unknown backend type: %s", typ)
	}

	if len(typs) > 1 {
		cfg.URI.Scheme = typs[1]
	}

	loggerName := fmt.Sprintf("backend.%s", typ)
	logger := zap.L().Named(loggerName).With(zap.Any("backend", cfg.URI))

	be, err := p.New(cfg, logger)
	if err != nil {
		return nil, err
	}

	return be, nil
}
