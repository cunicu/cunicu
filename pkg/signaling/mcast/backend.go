// SPDX-FileCopyrightText: 2025 Adam Rizkalla <ajarizzo@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package mcast

import (
	"context"
	"fmt"
	"net"
	"syscall"

	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/log"
	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
	"cunicu.li/cunicu/pkg/signaling"
	"go.uber.org/zap"
	"golang.org/x/net/ipv4"
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

	send_conn  net.PacketConn
	recv_conn  *net.UDPConn
	mcast_addr *net.UDPAddr
	config     BackendConfig

	logger *log.Logger
}

func NewBackend(cfg *signaling.BackendConfig, logger *log.Logger) (signaling.Backend, error) {
	b := &Backend{
		SubscriptionsRegistry: signaling.NewSubscriptionsRegistry(),
		logger:                logger,
	}

	//if err := b.config.Parse(cfg); err != nil {
	//	return nil, fmt.Errorf("failed to parse backend configuration: %w", err)
	//}

	var err error

	// Parse multicast group
	if b.mcast_addr, err = net.ResolveUDPAddr("udp", "224.0.0.1:9999"); err != nil {
		return nil, fmt.Errorf("Error parsing multicast address: %w", err)
	}

	// Bind to any available local UDP port for sending to multicast group
	if b.send_conn, err = net.ListenPacket("udp", ":0"); err != nil {
		return nil, fmt.Errorf("Error binding to local address: %w", err)
	}

	p := ipv4.NewPacketConn(b.send_conn)

	if err := p.JoinGroup(nil, b.mcast_addr); err != nil {
		return nil, fmt.Errorf("Error joining multicast group: %w", err)
	}

	// Add listener for multicast group
	if b.recv_conn, err = net.ListenMulticastUDP("udp", nil, b.mcast_addr); err != nil {
		return nil, fmt.Errorf("Error adding multicast listener: %w", err)
	}

	// Enable multicast loopback
	fd, _ := b.recv_conn.File()
	syscall.SetsockoptInt(int(fd.Fd()), syscall.IPPROTO_IP, syscall.IP_MULTICAST_LOOP, 1)
	//syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_MULTICAST_LOOP, 1)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, _, err := b.recv_conn.ReadFrom(buf)
			if err != nil {
				if err == net.ErrClosed {
					break
				}
				b.logger.Error("Error reading from UDPConn", zap.Error(err))
				break
				//continue
			}

			var env signalingproto.Envelope
			err = proto.Unmarshal(buf[:n], &env)
			if err != nil {
				b.logger.Error("Error unmarshaling protobuf", zap.Error(err))
				continue
			}

			if err := b.SubscriptionsRegistry.NewMessage(&env); err != nil {
				if err == signaling.ErrNotSubscribed {
					// Message wasn't for us but we will get everything over multicast, just ignore it
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

	if _, err = b.send_conn.WriteTo(data, b.mcast_addr); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (b *Backend) Close() error {
	//return fmt.Errorf("Close() called")
	//if err := b.conn.Close(); err != nil {
	//	return fmt.Errorf("failed to close multicast connection: %w", err)
	//}

	return nil
}
