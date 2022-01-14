package ice

import (
	"os"
	"unicode"

	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"riasc.eu/wice/internal/log"
)

type LoggerFactory struct {
	Base *zap.Logger
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

func (f *LoggerFactory) hook(e zapcore.Entry) error {
	runes := []rune(e.Message)

	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	e.Message = string(runes)

	return nil
}

func (f *LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	levelStr := os.Getenv("PION_LOG")

	var level zapcore.Level
	level.UnmarshalText([]byte(levelStr))

	logger := f.Base.Named("ice").WithOptions(
		zap.Hooks(f.hook),
		zap.Fields(zap.String("scope", scope)),
		log.WithLevel(level),
	)

	return &LeveledLogger{
		logger: logger.Sugar(),
	}
}
