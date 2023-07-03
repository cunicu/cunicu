// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"regexp"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
)

var grpcLogScope = regexp.MustCompile(`(?m)^\[(\w+)\]$`)

var _ grpclog.LoggerV2 = (*grpcLogger)(nil)

type grpcLogger struct {
	*zap.SugaredLogger
}

func (l *grpcLogger) Warning(args ...any) {
	l.SugaredLogger.Warn(args...)
}

func (l *grpcLogger) Warningln(args ...any) {
	l.SugaredLogger.Warnln(args...)
}

func (l *grpcLogger) Warningf(format string, args ...any) {
	l.SugaredLogger.Warnf(format, args...)
}

func (l *grpcLogger) Info(args ...any) {
	l.log(TraceLevel, "", args)
}

func (l *grpcLogger) Infof(format string, args ...any) {
	l.log(TraceLevel, format, args)
}

func (l *grpcLogger) InfoDepth(_ int, args ...any) {
	l.log(TraceLevel, "", args)
}

func (l *grpcLogger) WarningDepth(_ int, args ...any) {
	l.log(WarnLevel, "", args)
}

func (l *grpcLogger) ErrorDepth(_ int, args ...any) {
	l.log(ErrorLevel, "", args)
}

func (l *grpcLogger) FatalDepth(_ int, args ...any) {
	l.log(FatalLevel, "", args)
}

func (l *grpcLogger) unwrap(args []any) (string, []any) {
	scope := ""

	if len(args) > 0 {
		if str, ok := args[0].(string); ok {
			if m := grpcLogScope.FindStringSubmatch(str); m != nil {
				scope = m[1]
				args = args[1:]
			}
		}
	}

	return scope, args
}

func (l *grpcLogger) log(lvl Level, format string, args []any) {
	scope, args := l.unwrap(args)

	d := l.Desugar().Named(scope)

	// We check to avoid an unnecessary call to fmt.Sprint()
	if d.Check(zapcore.Level(lvl), "") != nil {
		if format != "" {
			d.Log(zapcore.Level(lvl), fmt.Sprintf(format, args...))
		} else {
			d.Log(zapcore.Level(lvl), fmt.Sprint(args...))
		}
	}
}

func (l *grpcLogger) V(lvl int) bool {
	return lvl >= Level(l.Level()).Verbosity()
}

func NewGRPCLogger(logger *Logger) grpclog.LoggerV2 {
	return &grpcLogger{
		SugaredLogger: logger.Sugar(),
	}
}
