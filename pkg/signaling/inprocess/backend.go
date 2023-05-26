// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package inprocess

import (
	"context"

	"github.com/stv0g/cunicu/pkg/crypto"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
	"github.com/stv0g/cunicu/pkg/signaling"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals
var subs = signaling.NewSubscriptionsRegistry()

func init() { //nolint:gochecknoinits
	signaling.Backends["inprocess"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "In-Process",
	}
}

type Backend struct {
	logger *zap.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *zap.Logger) (signaling.Backend, error) {
	b := &Backend{
		logger: logger,
	}

	for _, h := range cfg.OnReady {
		h.OnSignalingBackendReady(b)
	}

	return b, nil
}

func (b *Backend) Type() signalingproto.BackendType {
	return signalingproto.BackendType_INPROCESS
}

func (b *Backend) Subscribe(_ context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	return subs.Subscribe(kp, h)
}

func (b *Backend) Unsubscribe(_ context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	return subs.Unsubscribe(kp, h)
}

func (b *Backend) Publish(_ context.Context, kp *crypto.KeyPair, msg *signaling.Message) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return err
	}

	return subs.NewMessage(env)
}

func (b *Backend) Close() error {
	return nil
}
