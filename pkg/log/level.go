// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"bytes"
	"strconv"

	"go.uber.org/zap/zapcore"
)

func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	m := Level(l)
	enc.AppendString(m.String())
}

type Level zapcore.Level //nolint:recvcheck

const (
	DebugLevel Level = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

//nolint:gochecknoglobals
var (
	MaxLevel   = FatalLevel
	MinLevel   = VerboseLevel(10)
	TraceLevel = MinLevel
)

func VerboseLevel(v int) Level {
	return DebugLevel - Level(v) //nolint:gosec
}

func (l *Level) UnmarshalText(text []byte) error {
	if bytes.HasPrefix(text, []byte("debug")) {
		var v int

		if vs := string(text[5:]); len(vs) > 0 {
			v, _ = strconv.Atoi(vs)
		}

		*l = DebugLevel - Level(v) //nolint:gosec

		return nil
	}

	ll, err := zapcore.ParseLevel(string(text))
	if err != nil {
		return err
	}

	*l = Level(ll)

	return nil
}

func (l Level) String() string {
	if l < DebugLevel {
		return zapcore.DebugLevel.String() + strconv.Itoa(l.Verbosity())
	}

	return zapcore.Level(l).String()
}

func (l Level) Verbosity() int {
	return -int(l) - 1
}
