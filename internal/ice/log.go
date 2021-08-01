package ice

import (
	"strings"

	"github.com/pion/logging"
	log "github.com/sirupsen/logrus"
)

type LoggerFactory struct {
}

type Logger struct {
	log.Entry
}

func capitalize(msg string) string {
	for i, v := range msg {
		return strings.ToUpper(string(v)) + msg[i+1:]
	}
	return ""
}

func (l *Logger) Debug(msg string) {
	l.Entry.Debug(capitalize(msg))
}

func (l *Logger) Error(msg string) {
	l.Entry.Error(capitalize(msg))
}

func (l *Logger) Info(msg string) {
	l.Entry.Info(capitalize(msg))
}

func (l *Logger) Trace(msg string) {
	l.Entry.Trace(capitalize(msg))
}

func (l *Logger) Warn(msg string) {
	l.Entry.Warn(capitalize(msg))
}

func (f *LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	logger := &Logger{
		Entry: *log.WithFields(log.Fields{
			"logger": scope,
		}),
	}

	return logger
}
