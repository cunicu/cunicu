package log

import (
	"sync/atomic"
)

type VerbosityLevel struct {
	l atomic.Int32
}

func NewVerbosityLevel() VerbosityLevel {
	return VerbosityLevel{
		l: atomic.Int32{},
	}
}

func NewVerbosityLevelAt(l int) VerbosityLevel {
	a := NewVerbosityLevel()
	a.SetLevel(l)
	return a //nolint:govet
}

func (lvl VerbosityLevel) Enabled(l int) bool { //nolint:govet
	return lvl.Level() >= l || l == 0
}

func (lvl VerbosityLevel) Level() int { //nolint:govet
	return int(lvl.l.Load())
}

// SetLevel alters the logging level.
func (lvl VerbosityLevel) SetLevel(l int) { //nolint:govet
	lvl.l.Store(int32(l))
}