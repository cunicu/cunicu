// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/proto"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
	"github.com/stv0g/cunicu/pkg/signaling"
)

func init() { //nolint:gochecknoinits
	signaling.Backends["grpc"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "gRPC",
	}
}

type Backend struct {
	signaling.SubscriptionsRegistry

	client signalingproto.SignalingClient
	conn   *grpc.ClientConn

	config BackendConfig

	logger *log.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *log.Logger) (signaling.Backend, error) {
	var err error

	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),
		logger:                logger,
	}

	if err := b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	}

	// TODO: Use DialWithContext
	if b.conn, err = grpc.Dial(b.config.Target, b.config.Options...); err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	b.client = signalingproto.NewSignalingClient(b.conn)

	bi, err := b.client.GetBuildInfo(context.Background(), &proto.Empty{}, grpc.WaitForReady(true))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GRPC signaling server: %w", err)
	}

	b.logger.Debug("Connected to GRPC signaling server",
		zap.String("server_arch", bi.Arch),
		zap.String("server_version", bi.Version),
		zap.String("server_commit", bi.Commit),
		zap.String("server_tag", bi.Tag),
		zap.String("server_branch", bi.Branch),
		zap.String("server_os", bi.Os),
	)

	for _, h := range cfg.OnReady {
		h.OnSignalingBackendReady(b)
	}

	return b, nil
}

func (b *Backend) Type() signalingproto.BackendType {
	return signalingproto.BackendType_GRPC
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	first, err := b.SubscriptionsRegistry.Subscribe(kp, h)
	if err != nil {
		return false, err
	} else if first {
		pk := kp.Ours.PublicKey()
		return first, b.subscribeFromServer(ctx, &pk)
	}

	return first, nil
}

// Unsubscribe from messages send by a specific peer
func (b *Backend) Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	last, err := b.SubscriptionsRegistry.Unsubscribe(kp, h)
	if err != nil {
		return false, err
	} else if last {
		pk := kp.Ours.PublicKey()
		return last, b.unsubscribeFromServer(ctx, &pk)
	}

	return last, nil
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *signaling.Message) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	if _, err = b.client.Publish(ctx, env, grpc.WaitForReady(true)); err != nil {
		if status.Code(err) == codes.Canceled {
			return signaling.ErrClosed
		}

		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	if err := b.conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}

func (b *Backend) subscribeFromServer(ctx context.Context, pk *crypto.Key) error {
	params := &signalingproto.SubscribeParams{
		Key: pk.Bytes(),
	}

	stream, err := b.client.Subscribe(ctx, params, grpc.WaitForReady(true))
	if err != nil {
		return fmt.Errorf("failed to subscribe to offers: %w", err)
	}

	// Wait until subscription has been created
	// This avoids a race between Subscribe() / Publish() when two subscribers are subscribing
	// to each other.
	if _, err := stream.Recv(); err != nil {
		return fmt.Errorf("failed receive synchronization envelope: %w", err)
	}

	b.logger.Debug("Created new subscription", zap.Any("pk", pk))

	go func() {
		for {
			env, err := stream.Recv()
			if err != nil {
				if !errors.Is(err, io.EOF) && status.Code(err) != codes.Canceled {
					b.logger.Error("Subscription stream closed. Re-subscribing..", zap.Error(err))

					if err := b.subscribeFromServer(ctx, pk); err != nil && status.Code(err) != codes.Canceled {
						b.logger.Error("Failed to resubscribe", zap.Error(err))
					}
				}

				break
			}

			if err := b.SubscriptionsRegistry.NewMessage(env); err != nil {
				b.logger.Error("Failed to decrypt message", zap.Error(err))
			}
		}
	}()

	return nil
}

func (b *Backend) unsubscribeFromServer(_ context.Context, _ *crypto.Key) error {
	// TODO: Cancel subscription stream

	return nil
}
