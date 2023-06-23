// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-FileCopyrightText: 2020 Manfred Touron <https://manfred.life>
// SPDX-License-Identifier: Apache-2.0

package log

// This filter implements a filtering zap core
// Based on: https://github.com/moul/zapfilter

import (
	"errors"

	"go.uber.org/zap/zapcore"
)

var (
	ErrUnsupportedKeyword = errors.New("unsupported keyword")
	ErrBadSyntax          = errors.New("bad syntax")
)

// NewFilteredCore returns a core middleware that uses the given filter function to
// determine whether to actually call Write on the next core in the chain.
func NewFilteredCore(next zapcore.Core, rule *AtomicFilterRule) zapcore.Core {
	return &filteringCore{next, rule}
}

type filteringCore struct {
	next zapcore.Core
	rule *AtomicFilterRule
}

// Check determines whether the supplied zapcore.Entry should be logged.
// If the entry should be logged, the filteringCore adds itself to the zapcore.CheckedEntry
// and returns the results.
func (core *filteringCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if r := core.rule.Load(); r == nil || r.Filter(entry) {
		ce = ce.AddCore(entry, core)
	}
	return ce
}

// Write determines whether the supplied zapcore.Entry with provided []zapcore.Field should
// be logged, then calls the wrapped zapcore.Write.
func (core *filteringCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return core.next.Write(entry, fields)
}

// With adds structured context to the wrapped zapcore.Core.
func (core *filteringCore) With(fields []zapcore.Field) zapcore.Core {
	return &filteringCore{
		next: core.next.With(fields),
		rule: core.rule,
	}
}

// Enabled asks the wrapped zapcore.Core to decide whether a given logging level is enabled
// when logging a message.
func (core *filteringCore) Enabled(_ zapcore.Level) bool {
	return true
}

func (core *filteringCore) Level() zapcore.Level {
	if r := core.rule.Load(); r != nil {
		return zapcore.Level(r.Level)
	}
	return zapcore.InvalidLevel
}

// Sync flushed buffered logs (if any).
func (core *filteringCore) Sync() error {
	return core.next.Sync()
}
