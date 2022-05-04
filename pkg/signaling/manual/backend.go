package manual

import (
	"context"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

func init() {
	signaling.Backends["manual"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "Manual",
	}
}

type Backend struct {
	signaling.Backend

	logger *zap.Logger
	events chan *pb.Event

	lastMessage *pb.EncryptedMessage
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

func (b *Backend) Close() error {
	return nil
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	b.logger.Error("TODO: publish")
	return nil
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	b.logger.Error("TODO: subscribe")
	return nil, nil
}
