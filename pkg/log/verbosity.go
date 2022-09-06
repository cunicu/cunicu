package log

import (
	"go.uber.org/atomic"
)

type VerbosityLevel struct {
	l *atomic.Int32
}

func NewVerbosityLevel() VerbosityLevel {
	return VerbosityLevel{
		l: atomic.NewInt32(0),
	}
}

func NewVerbosityLevelAt(l int) VerbosityLevel {
	a := NewVerbosityLevel()
	a.SetLevel(l)
	return a
}

func (lvl VerbosityLevel) Enabled(l int) bool {
	return lvl.Level() >= l || l == 0
}

func (lvl VerbosityLevel) Level() int {
	return int(lvl.l.Load())
}

// SetLevel alters the logging level.
func (lvl VerbosityLevel) SetLevel(l int) {
	lvl.l.Store(int32(l))
}
