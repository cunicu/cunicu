package ice

import (
	"fmt"
	"os"
	"unicode"

	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerFactory struct {
	Base *zap.Logger
}

type LeveledLogger struct {
	logging.LeveledLogger

	*zap.SugaredLogger
}

func (l *LeveledLogger) Trace(msg string) {
	l.SugaredLogger.Debug(msg)
}

func (l *LeveledLogger) Tracef(format string, args ...interface{}) {
	l.SugaredLogger.Debugf(format, args...)
}

func (l *LeveledLogger) Debug(msg string) {
	l.SugaredLogger.Debug(msg)
}

func (l *LeveledLogger) Debugf(format string, args ...interface{}) {
	l.SugaredLogger.Debugf(format, args...)
}

func (l *LeveledLogger) Info(msg string) {
	l.SugaredLogger.Info(msg)
}

func (l *LeveledLogger) Infof(format string, args ...interface{}) {
	l.SugaredLogger.Infof(format, args...)
}

func (l *LeveledLogger) Warn(msg string) {
	l.SugaredLogger.Warn(msg)
}

func (l *LeveledLogger) Warnf(format string, args ...interface{}) {
	l.SugaredLogger.Warnf(format, args...)
}

func (l *LeveledLogger) Error(msg string) {
	l.SugaredLogger.Error(msg)
}

func (l *LeveledLogger) Errorf(format string, args ...interface{}) {
	l.SugaredLogger.Errorf(format, args...)
}

func (f *LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	var lvl zapcore.Level
	if lvlStr := os.Getenv("PION_LOG"); lvlStr != "" {
		lvl.UnmarshalText([]byte(lvlStr))
	} else {
		lvl = zapcore.DebugLevel
	}

	loggerName := "ice"
	if scope != "ice" {
		loggerName += fmt.Sprintf(".%s", scope)
	}

	logger := f.Base.Named(loggerName).WithOptions(
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return &pionCore{
				Core:      c,
				PionLevel: lvl,
			}
		}),
	)

	return &LeveledLogger{
		SugaredLogger: logger.Sugar(),
	}
}

func NewLogger(base *zap.Logger, scope string) logging.LeveledLogger {
	lf := LoggerFactory{Base: base}
	return lf.NewLogger(scope)
}

type pionCore struct {
	zapcore.Core
	PionLevel zapcore.Level
}

func (c *pionCore) Write(e zapcore.Entry, f []zapcore.Field) error {
	runes := []rune(e.Message)

	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	e.Message = string(runes)

	return c.Core.Write(e, f)
}

func (c *pionCore) Check(e zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.PionLevel.Enabled(e.Level) {
		return ce.AddCore(e, c)
	}

	return ce
}
