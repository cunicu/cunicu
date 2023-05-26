// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"net"

	"github.com/pion/stun"
	"go.uber.org/zap"
)

type STUNPacketHandler struct {
	Logger *zap.Logger
}

func (ph *STUNPacketHandler) OnPacketRead(buf []byte, rAddr net.Addr) (bool, error) {
	if !stun.IsMessage(buf) {
		return false, nil
	}

	if ph.Logger != nil {
		msg := &stun.Message{
			Raw: buf,
		}

		if err := msg.Decode(); err == nil {
			ph.Logger.Debug("Received STUN message",
				zap.String("addr", rAddr.String()),
				zap.Any("type", msg.Type),
				zap.Binary("id", msg.TransactionID[:]),
				zap.Int("#attrs", len(msg.Attributes)),
				zap.Int("len", int(msg.Length)))
		} else {
			ph.Logger.Debug("Received invalid STUN message",
				zap.String("addr", rAddr.String()),
				zap.Int("len", len(buf)),
				zap.Binary("msg", buf))
		}
	}

	return true, nil
}
