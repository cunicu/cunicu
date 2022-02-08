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
	logger *zap.Logger
	events chan *pb.Event
}

func NewBackend(cfg *signaling.BackendConfig, events chan *pb.Event, logger *zap.Logger) (signaling.Backend, error) {
	b := &Backend{
		events: events,
		logger: logger,
	}

	b.events <- &pb.Event{
		Type: pb.Event_BACKEND_READY,
		Event: &pb.Event_BackendReady{
			BackendReady: &pb.BackendReadyEvent{
				Type: pb.BackendReadyEvent_INPROCESS,
			},
		},
	}

	return b, nil
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	sub, err := subs.NewSubscription(kp)
	if err != nil {
		return nil, err
	}

	return sub.Channel, nil
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
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
