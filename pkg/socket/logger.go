package socket

import (
	"regexp"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/grpclog"
)

var grpcLogExpr = regexp.MustCompile(`(?m)^\[(\w+)\] (.*)$`)

type grpcLogHook struct{}

func (h *grpcLogHook) Fire(e *logrus.Entry) error {
	if m := grpcLogExpr.FindStringSubmatch(e.Message); m != nil {
		e.Data["source"] = m[1]
		e.Message = m[2]
	}

	return nil
}

func (h *grpcLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type Logger struct {
	logrus.FieldLogger

	Level int
}

func NewLogger(lvl int) grpclog.LoggerV2 {
	l := logrus.WithField("logger", "grpc")

	l.Logger.AddHook(&grpcLogHook{})

	return &Logger{
		FieldLogger: l,
		Level:       lvl,
	}
}

func (l *Logger) V(lvl int) bool {
	return lvl > l.Level
}

func init() {
	l := NewLogger(0)

	grpclog.SetLoggerV2(l)
}
