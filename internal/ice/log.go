package ice

import (
	"unicode"

	"github.com/pion/logging"
	log "github.com/sirupsen/logrus"
)

type LoggerFactory struct {
}

type Logger struct {
	*log.Entry
}

func capitalize(msg string) string {
	runes := []rune(msg)

	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return string(runes)
}

func (l *Logger) Debug(msg string) {
	msg = capitalize(msg)
	l.Entry.Debug(msg)
}

func (l *Logger) Error(msg string) {
	msg = capitalize(msg)
	l.Entry.Error(msg)
}

func (l *Logger) Info(msg string) {
	msg = capitalize(msg)
	l.Entry.Info(msg)
}

func (l *Logger) Trace(msg string) {
	msg = capitalize(msg)
	l.Entry.Trace(msg)
}

func (l *Logger) Warn(msg string) {
	msg = capitalize(msg)
	l.Entry.Warn(msg)
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	format = capitalize(format)
	l.Entry.Tracef(format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	format = capitalize(format)
	l.Entry.Debugf(format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	format = capitalize(format)
	l.Entry.Infof(format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	format = capitalize(format)
	l.Entry.Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	format = capitalize(format)
	l.Entry.Errorf(format, args...)
}

func (f *LoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	return &Logger{
		Entry: log.WithFields(log.Fields{
			"logger": "ice",
			"scope":  scope,
		}),
	}
}
