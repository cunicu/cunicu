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
	signaling.SubscriptionsRegistry

	client pb.SignalingClient
	conn   *grpc.ClientConn

	isClosing bool

	config BackendConfig

	logger *zap.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *zap.Logger) (signaling.Backend, error) {
	var err error

	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),
		logger:                logger,
	}

	if err := b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	}

	if b.conn, err = grpc.Dial(b.config.Target, b.config.Options...); err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	b.client = pb.NewSignalingClient(b.conn)

	for _, h := range cfg.OnReady {
		h.OnSignalingBackendReady(b)
	}

	return b, nil
}

func (b *Backend) Type() pb.BackendType {
	return pb.BackendType_GRPC
}

func (b *Backend) SubscribeAll(ctx context.Context, sk *crypto.Key, h signaling.MessageHandler) error {
	if created, err := b.SubscriptionsRegistry.SubscribeAll(sk, h); err != nil {
		return err
	} else if created {
		pk := sk.PublicKey()
		return b.subscribeFromServer(ctx, &pk)
	}

	return nil
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) error {
	if created, err := b.SubscriptionsRegistry.Subscribe(kp, h); err != nil {
		return err
	} else if created {
		pk := kp.Ours.PublicKey()
		return b.subscribeFromServer(ctx, &pk)
	}

	return nil
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *signaling.Message) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	if _, err = b.client.Publish(ctx, env); err != nil {
		return fmt.Errorf("failed to publish offer: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	b.isClosing = true

	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (b *Backend) subscribeFromServer(ctx context.Context, pk *crypto.Key) error {
	b.logger.Debug("Creating new subscription", zap.Any("pk", pk))

	params := &pb.SubscribeParams{
		Key: pk.Bytes(),
	}

	stream, err := b.client.Subscribe(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to subscribe to offers: %s", err)
	}

	go func() {
		for {
			if env, err := stream.Recv(); err == nil {
				if err := b.SubscriptionsRegistry.NewMessage(env); err != nil {
					b.logger.Error("Failed to decrypt message", zap.Error(err))
				}
			} else {
				if b.isClosing {
					return
				}

				b.logger.Error("Failed to receive offer", zap.Error(err))
			}
		}
	}()

	return nil
}
