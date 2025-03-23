// SPDX-FileCopyrightText: 2025 Adam Rizkalla <ajarizzo@gmail.com>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

// Package mcast implements a signaling backend using multicast
package mcast

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
	"cunicu.li/cunicu/pkg/signaling"
)

var errInvalidAddress = errors.New("missing multicast address")

func ParseURL(urlStr string) (string, BackendOptions, error) {
	o := BackendOptions{
		Interface: nil,
		Loopback:  false,
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", o, err
	}

	q := u.Query()

	if q.Has("interface") {
		if o.Interface, err = net.InterfaceByName(q.Get("interface")); err != nil {
			return "", o, fmt.Errorf("failed to parse 'interface' option: %w", err)
		}
	}

	if q.Has("loopback") {
		var err error
		if o.Loopback, err = strconv.ParseBool(q.Get("loopback")); err != nil {
			return "", o, fmt.Errorf("failed to parse 'loopback' option: %w", err)
		}
	}

	if u.Host == "" {
		return "", o, errInvalidAddress
	}

	return u.Host, o, nil
}

func (b *Backend) run() {
	buf := make([]byte, 4096)

	for {
		n, err := b.conn.Read(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
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
			if !errors.Is(err, signaling.ErrNotSubscribed) {
				b.logger.Error("Failed to decrypt message", zap.Error(err))
			}

			continue
		}
	}
}
