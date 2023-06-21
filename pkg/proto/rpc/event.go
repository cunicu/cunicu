// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"strings"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/log"
)

func (e *Event) Log(l *log.Logger, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("type", strings.ToLower(e.Type.String())))

	if e.Event != nil {
		fields = append(fields, zap.Any("event", e.Event))
	}

	if e.Interface != "" {
		fields = append(fields, zap.String("interface", e.Interface))
	}

	if e.Peer != nil {
		pk, err := crypto.ParseKeyBytes(e.Peer)
		if err != nil {
			panic(err)
		}

		fields = append(fields, zap.Any("peer", pk))
	}

	if e.Time != nil {
		fields = append(fields, zap.Time("time", e.Time.Time()))
	}

	l.Info(msg, fields...)
}
