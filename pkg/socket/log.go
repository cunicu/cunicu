package socket

import (
	"fmt"
	"regexp"

	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

var grpcLogExpr = regexp.MustCompile(`(?m)^\[(\w+)\]$`)

type grpcLogger struct {
	*zap.SugaredLogger
	verbosity int
}

func NewLogger(logger *zap.Logger, verbosity int) grpclog.LoggerV2 {
	return &grpcLogger{
		SugaredLogger: logger.Sugar(),
		verbosity:     verbosity,
	}
}

func (l *grpcLogger) unwrap(args []interface{}) (string, []zap.Field) {
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

func (l *grpcLogger) Warning(args ...interface{}) {
	l.Warn(args...)
}

func (l *grpcLogger) Warningf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func (l *grpcLogger) Infoln(args ...interface{}) {
	msg, fields := l.unwrap(args)
	l.Desugar().Info(msg, fields...)
}

func (l *grpcLogger) Warningln(args ...interface{}) {
	msg, fields := l.unwrap(args)
	l.Desugar().Warn(msg, fields...)
}

func (l *grpcLogger) Errorln(args ...interface{}) {
	msg, fields := l.unwrap(args)
	l.Desugar().Error(msg, fields...)
}

func (l *grpcLogger) Fatalln(args ...interface{}) {
	msg, fields := l.unwrap(args)
	l.Desugar().Fatal(msg, fields...)
}

func (l *grpcLogger) V(lvl int) bool {
	return lvl > l.verbosity
}
