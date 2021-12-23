package ice

import (
	"unicode"

	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LoggerFactory struct {
	base *zap.Logger
}

type LeveledLogger struct {
	logging.LeveledLogger

	logger *zap.SugaredLogger
}

func (l *LeveledLogger) Trace(msg string) {
	l.logger.Debug(msg)
}

func (l *LeveledLogger) Tracef(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LeveledLogger) Debug(msg string) {
	l.logger.Debug(msg)
}

func (l *LeveledLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debugf(format, args...)
}

func (l *LeveledLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *LeveledLogger) Infof(format string, args ...interface{}) {
	l.logger.Infof(format, args...)
}

func (l *LeveledLogger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *LeveledLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}

func (l *LeveledLogger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *LeveledLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}

func capitalize(msg string) string {
	runes := []rune(msg)

	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return string(runes)
}

func (f *LoggerFactory) hook(e zapcore.Entry) error {
	e.Message = capitalize(e.Message)
	return nil
}

func (f *LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	logger := f.base.Named("ice").WithOptions(
		zap.Hooks(f.hook),
		zap.Fields(zap.String("scope", scope)),
	)

	return &LeveledLogger{
		logger: logger.Sugar(),
	}
}
