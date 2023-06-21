// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"net"

	"github.com/pion/stun"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/log"
)

type STUNPacketHandler struct {
	Logger *log.Logger
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
			ph.Logger.DebugV(6, "Received STUN message",
				zap.String("addr", rAddr.String()),
				zap.Any("type", msg.Type),
				zap.Binary("id", msg.TransactionID[:]),
				zap.Int("#attrs", len(msg.Attributes)),
				zap.Int("len", int(msg.Length)))
		} else {
			ph.Logger.DebugV(6, "Received invalid STUN message",
				zap.String("addr", rAddr.String()),
				zap.Int("len", len(buf)),
				zap.Binary("msg", buf))
		}
	}

	return true, nil
}
