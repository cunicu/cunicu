// SPDX-FileCopyrightText: 2025 Adam Rizkalla <ajarizzo@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package mcast

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"

	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/log"
	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
	"cunicu.li/cunicu/pkg/signaling"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func init() { //nolint:gochecknoinits
	signaling.Backends["multicast"] = &signaling.BackendPlugin{
		New:         NewBackend,
		Description: "Multicast",
	}
}

type Backend struct {
	signaling.SubscriptionsRegistry

	conn      *net.UDPConn
	mcastAddr *net.UDPAddr
	config    BackendConfig

	logger *log.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *log.Logger) (signaling.Backend, error) {
	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),
		logger:                logger,
	}

	var err error

	if err = b.config.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	}

	// Parse multicast group address.
	if b.mcastAddr, err = net.ResolveUDPAddr("udp", b.config.Target); err != nil {
		return nil, fmt.Errorf("Error parsing multicast address: %w", err)
	}

	// Add listener for multicast group.
	if b.conn, err = net.ListenMulticastUDP("udp", b.config.Options.Interface, b.mcastAddr); err != nil {
		return nil, fmt.Errorf("Error adding multicast listener: %w", err)
	}

	if b.config.Options.Loopback {
		// Enable multicast loopback.
		fd, _ := b.conn.File()
		syscall.SetsockoptInt(int(fd.Fd()), syscall.IPPROTO_IP, syscall.IP_MULTICAST_LOOP, 1)
	}

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := b.conn.Read(buf)
			if err != nil {
				if err == net.ErrClosed {
					break
				}
				b.logger.Error("Error reading from UDPConn", zap.Error(err))
				continue
			}

			var env signalingproto.Envelope
			if err = proto.Unmarshal(buf[:n], &env); err != nil {
				b.logger.Error("Error unmarshaling protobuf", zap.Error(err))
				continue
			}

			if err := b.SubscriptionsRegistry.NewMessage(&env); err != nil {
				if errors.Is(err, signaling.ErrNotSubscribed) {
					// Message wasn't for us but we will get everything over multicast, just ignore it.
				} else {
					b.logger.Error("Failed to decrypt message", zap.Error(err))
				}
				continue
			}
		}
	}()

	for _, h := range cfg.OnReady {
		h.OnSignalingBackendReady(b)
	}

	return b, nil
}

func (b *Backend) Type() signalingproto.BackendType {
	return signalingproto.BackendType_MCAST
}

func (b *Backend) Subscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	return b.SubscriptionsRegistry.Subscribe(kp, h)
}

func (b *Backend) Unsubscribe(ctx context.Context, kp *crypto.KeyPair, h signaling.MessageHandler) (bool, error) {
	return b.SubscriptionsRegistry.Unsubscribe(kp, h)
}

func (b *Backend) Publish(ctx context.Context, kp *crypto.KeyPair, msg *signaling.Message) error {
	env, err := msg.Encrypt(kp)
	if err != nil {
		return fmt.Errorf("failed to encrypt message: %w", err)
	}

	data, err := proto.Marshal(env)
	if err != nil {
		return fmt.Errorf("Error marshaling protobuf: %w", err)
	}

	if _, err = b.conn.WriteTo(data, b.mcastAddr); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	// NOTE: Do not close the connection; on certain OS (like Linux),
	// the UDPConn.Read() will continue to block even if the connection
	// is closed.
	//if err := b.conn.Close(); err != nil {
	//	return fmt.Errorf("failed to close multicast connection: %w", err)
	//}

	return nil
}
