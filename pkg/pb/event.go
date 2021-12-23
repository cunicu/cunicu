package pb

import (
	"go.uber.org/zap"
)

func (e *Event) Log(l *zap.Logger, msg string, fields ...zap.Field) {
	fields = append(fields,
		zap.String("type", e.Type),
		zap.String("state", e.State),
	)

	if e.Time != nil {
		fields = append(fields, zap.Any("time", e.Time.Time()))
	}

	l.Info(msg, fields...)
}

func (e *Event) Match(o *Event) bool {
	if e.Type != o.Type {
		return false
	}

	if e.State != o.State {
		return false
	}

	return true
}
