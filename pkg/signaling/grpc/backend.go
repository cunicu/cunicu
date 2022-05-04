package grpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

func init() {
	signaling.Backends["grpc"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "gRPC",
	}
}

type Backend struct {
	client pb.SignalingClient
	conn   *grpc.ClientConn

	config BackendConfig

	events chan *pb.Event
	logger *zap.Logger
}

func NewBackend(cfg *signaling.BackendConfig, events chan *pb.Event, logger *zap.Logger) (signaling.Backend, error) {
	var err error

	b := &Backend{
		events: events,
		logger: logger,
	}

	if err := b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	}

	if b.conn, err = grpc.Dial(b.config.Target, b.config.Options...); err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	b.client = pb.NewSignalingClient(b.conn)

	b.events <- &pb.Event{
		Type: pb.Event_BACKEND_READY,
		Event: &pb.Event_BackendReady{
			BackendReady: &pb.BackendReadyEvent{
				Type: pb.BackendReadyEvent_GRPC,
			},
		},
	}

	return b, nil
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair) (chan *pb.SignalingMessage, error) {
	params := &pb.SubscribeParams{
		Key: kp.Ours.PublicKey().Bytes(),
	}

	stream, err := b.client.Subscribe(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to offers: %s", err)
	}

	ch := make(chan *pb.SignalingMessage)

	go func() {
		for {
			if env, err := stream.Recv(); err == nil {
				if msg, err := env.Decrypt(kp); err == nil {
					ch <- msg
				} else {
					b.logger.Error("Failed to decrypt message", zap.Error(err))
				}
			} else {
				b.logger.Error("Failed to receive offer", zap.Error(err))
			}
		}
	}()

	return ch, nil
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *pb.SignalingMessage) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to encryt message: %w", err)
	}

	if _, err = b.client.Publish(ctx, env); err != nil {
		return fmt.Errorf("failed to publish offer: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}
