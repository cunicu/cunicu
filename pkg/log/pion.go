// SPDX-FileCopyrightText: 2023 Atsushi Watanabe <atsushi.w@ieee.org>
// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: MIT

package log

import (
	"fmt"
	"sync"

	"github.com/pion/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	_ logging.LeveledLogger = (*pionLogger)(nil)
	_ logging.LoggerFactory = (*PionLoggerFactory)(nil)
)

type pionLogger struct {
	*zap.SugaredLogger
}

func (l *pionLogger) Trace(msg string) {
	l.Desugar().Log(zapcore.Level(TraceLevel), msg)
}

func (l *pionLogger) Tracef(format string, args ...interface{}) {
	// Check for level before calling fmt.Sprintf()
	if d := l.Desugar(); d.Check(zapcore.Level(TraceLevel), "") != nil {
		d.Log(zapcore.Level(TraceLevel), fmt.Sprintf(format, args...))
	}
}

func (l *pionLogger) Debug(msg string) {
	l.SugaredLogger.Debug(msg)
}

func (l *pionLogger) Info(msg string) {
	l.SugaredLogger.Info(msg)
}

func (l *pionLogger) Warn(msg string) {
	l.SugaredLogger.Warn(msg)
}

func (l *pionLogger) Error(msg string) {
	l.SugaredLogger.Error(msg)
}

// PionLoggerFactory is a logger factory backended by zap logger.
type PionLoggerFactory struct {
	base *zap.Logger

	mu      sync.Mutex
	loggers []*pionLogger
}

// NewLogger creates new scoped logger.
func (f *PionLoggerFactory) NewLogger(scope string) logging.LeveledLogger {
	f.mu.Lock()
	defer f.mu.Unlock()

	l := &pionLogger{
		SugaredLogger: f.base.Named(scope).WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
	f.loggers = append(f.loggers, l)

	return l
}

// SyncAll calls Sync() method of all child loggers.
// It is recommended to call this before exiting the program to
// ensure never loosing buffered log.
func (f *PionLoggerFactory) SyncAll() {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, l := range f.loggers {
		_ = l.SugaredLogger.Sync()
	}
}

func NewPionLoggerFactory(base *Logger) logging.LoggerFactory {
	return &PionLoggerFactory{
		base: base.Logger,
	}
}

func NewPionLogger(base *Logger, scope string) logging.LeveledLogger {
	return NewPionLoggerFactory(base).NewLogger(scope)
}
