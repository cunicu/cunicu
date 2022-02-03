package inprocess

import (
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
	logger *zap.Logger
}

func NewBackend(cfg *signaling.BackendConfig, events chan *pb.Event) (signaling.Backend, error) {
	b := &Backend{
		logger: zap.L().Named("in-process"),
	}

	return b, nil
}

func (b *Backend) Subscribe(kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	sub, err := subs.NewSubscription(kp)
	if err != nil {
		return nil, err
	}

	return sub.Channel, nil
}

func (b *Backend) Publish(kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	_, err := subs.NewSubscription(kp)
	if err != nil {
		return err
	}

	env, err := msg.Encrypt(kp)
	if err != nil {
		return err
	}

	return subs.NewMessage(env)
}

func (b *Backend) Close() error {

	return nil
}
