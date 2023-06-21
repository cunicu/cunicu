// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func (l *Logger) Level() Level {
	return Level(l.Logger.Level())
}

func (l *Logger) DebugV(verboseLevel int, message string, fields ...zap.Field) {
	l.Log(zap.DebugLevel-zapcore.Level(verboseLevel), message, fields...)
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

func (l *Logger) WithOptions(opts ...zap.Option) *Logger {
	return &Logger{l.Logger.WithOptions(opts...)}
}

func (l *Logger) Named(s string) *Logger {
	return &Logger{l.Logger.Named(s)}
}
