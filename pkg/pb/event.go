package pb

import (
	"strings"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
)

func (e *Event) Log(l *zap.Logger, msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("type", strings.ToLower(e.Type.String())),
		zap.Any("event", e.Event),
	)

	if e.Interface != "" {
		fields = append(fields, zap.String("interface", e.Interface))
	}

	if e.Peer != nil {
		pk := (*crypto.Key)(e.Peer)
		fields = append(fields, zap.Any("peer", pk))
	}

	if e.Time != nil {
		fields = append(fields, zap.Time("time", e.Time.Time()))
	}

	l.Info(msg, fields...)
}
