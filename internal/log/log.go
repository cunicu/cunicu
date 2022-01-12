package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// WithLevel sets the loggers leveler to the one given.
func WithLevel(level zapcore.LevelEnabler) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &lvlCore{Core: core, l: level}
	})
}

type lvlCore struct {
	zapcore.Core
	l zapcore.LevelEnabler
}

func (c *lvlCore) Enabled(lvl zapcore.Level) bool {
	return c.l.Enabled(lvl)
}

func (c *lvlCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.l.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}
