package log

import (
	"fmt"
	"os"
	"unicode"

	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type pionLoggerFactory struct {
	Base *zap.Logger
}

type pionLeveledLogger struct {
	logging.LeveledLogger

	*zap.SugaredLogger
}

func (l *pionLeveledLogger) Trace(msg string) {
	l.SugaredLogger.Debug(msg)
}

func (l *pionLeveledLogger) Tracef(format string, args ...any) {
	l.SugaredLogger.Debugf(format, args...)
}

func (l *pionLeveledLogger) Debug(msg string) {
	l.SugaredLogger.Debug(msg)
}

func (l *pionLeveledLogger) Debugf(format string, args ...any) {
	l.SugaredLogger.Debugf(format, args...)
}

func (l *pionLeveledLogger) Info(msg string) {
	l.SugaredLogger.Info(msg)
}

func (l *pionLeveledLogger) Infof(format string, args ...any) {
	l.SugaredLogger.Infof(format, args...)
}

func (l *pionLeveledLogger) Warn(msg string) {
	l.SugaredLogger.Warn(msg)
}

func (l *pionLeveledLogger) Warnf(format string, args ...any) {
	l.SugaredLogger.Warnf(format, args...)
}

func (l *pionLeveledLogger) Error(msg string) {
	l.SugaredLogger.Error(msg)
}

func (l *pionLeveledLogger) Errorf(format string, args ...any) {
	l.SugaredLogger.Errorf(format, args...)
}

func (f *pionLoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	var lvl zapcore.Level
	if lvlStr := os.Getenv("PION_LOG"); lvlStr != "" {
		if err := lvl.UnmarshalText([]byte(lvlStr)); err != nil {
			f.Base.Fatal("Unknown ICE logger level", zap.Error(err), zap.String("level", lvlStr))
		}
	} else {
		lvl = zapcore.WarnLevel
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

	return &pionLeveledLogger{
		SugaredLogger: logger.Sugar(),
	}
}

func NewPionLoggerFactory(base *zap.Logger) *pionLoggerFactory {
	return &pionLoggerFactory{Base: base}
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
