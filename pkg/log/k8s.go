package log

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type k8sCore struct {
	zapcore.Core
}

func (c *k8sCore) Write(e zapcore.Entry, f []zapcore.Field) error {
	// Strip newline at the end
	e.Message = strings.TrimSpace(e.Message)

	return c.Core.Write(e, f)
}

func (c *k8sCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return ce.AddCore(e, c)
	}

	return ce
}

func NewK8SLogger(base *zap.Logger) logr.Logger {
	base = base.WithOptions(
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return &k8sCore{
				Core: c,
			}
		}),
	)

	return zapr.NewLogger(base)
}
