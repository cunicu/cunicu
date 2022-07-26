package log

import "go.uber.org/zap/zapcore"

type k8sCore struct {
	zapcore.Core
	PionLevel zapcore.Level
}

func (c *k8sCore) Write(e zapcore.Entry, f []zapcore.Field) error {
	e.Message = e.Message[:len(e.Message)-1]

	return c.Core.Write(e, f)
}

func (c *k8sCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(e.Level) {
		return ce.AddCore(e, c)
	}

	return ce
}
