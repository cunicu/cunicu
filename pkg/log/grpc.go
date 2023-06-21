// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
)

var grpcLogExpr = regexp.MustCompile(`(?m)^\[(\w+)\]$`)

type grpcLogger struct {
	*zap.SugaredLogger
	verbosity int
}

func NewGRPCLogger(logger *Logger, verbosity int) grpclog.LoggerV2 {
	var level zapcore.Level

	verbosityLevel := os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")
	if vl, err := strconv.Atoi(verbosityLevel); err == nil {
		verbosity = vl
	}

	severityLevel := os.Getenv("GRPC_GO_LOG_SEVERITY_LEVEL")
	if severityLevel != "" {
		if err := level.UnmarshalText([]byte(severityLevel)); err != nil {
			logger.Fatal("Unknown gRPC logger severity level", zap.Error(err), zap.String("level", severityLevel))
		}
	} else {
		level = zap.WarnLevel
	}

	logger = logger.WithOptions(zap.IncreaseLevel(level))

	return &grpcLogger{
		SugaredLogger: logger.Logger.Sugar(),
		verbosity:     verbosity,
	}
}

func (l *grpcLogger) unwrap(args []any) (string, []zap.Field) {
	fields := []zap.Field{}

	if len(args) > 0 {
		if str, ok := args[0].(string); ok {
			if m := grpcLogExpr.FindStringSubmatch(str); m != nil {
				fields = append(fields, zap.String("scope", m[1]))
				args = args[1:]
			}
		}
	}

	return fmt.Sprint(args...), fields
}

func (l *grpcLogger) Warning(args ...any) {
	l.Warn(args...)
}

func (l *grpcLogger) Warningf(format string, args ...any) {
	l.Warnf(format, args...)
}

func (l *grpcLogger) Infoln(args ...any) {
	msg, fields := l.unwrap(args)
	l.Desugar().Info(msg, fields...)
}

func (l *grpcLogger) Warningln(args ...any) {
	msg, fields := l.unwrap(args)
	l.Desugar().Warn(msg, fields...)
}

func (l *grpcLogger) Errorln(args ...any) {
	msg, fields := l.unwrap(args)
	l.Desugar().Error(msg, fields...)
}

func (l *grpcLogger) Fatalln(args ...any) {
	msg, fields := l.unwrap(args)
	l.Desugar().Fatal(msg, fields...)
}

func (l *grpcLogger) V(lvl int) bool {
	return lvl < l.verbosity
}
