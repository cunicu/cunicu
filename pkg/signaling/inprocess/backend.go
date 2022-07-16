package inprocess

import (
	"context"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

var (
	subs = signaling.NewSubscriptionsRegistry()
)

func init() {
	signaling.Backends["inprocess"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "In-Process",
	}
}

type Backend struct {
	onReady signaling.BackendReadyHandlerList

	logger *zap.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *zap.Logger) (signaling.Backend, error) {
	b := &Backend{
		onReady: signaling.BackendReadyHandlerList{},
		logger:  logger,
	}

	b.onReady.Invoke(b)

	return b, nil
}

func (b *Backend) OnReady(h signaling.BackendReadyHandler) {
	b.onReady.Register(h)
}

func (b *Backend) Type() pb.BackendReadyEvent_Type {
	return pb.BackendReadyEvent_INPROCESS
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	return subs.Subscribe(kp)
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return err
	}

	return subs.NewMessage(env)
}

func (b *Backend) Close() error {
	return nil
}
