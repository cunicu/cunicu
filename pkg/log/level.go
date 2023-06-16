// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
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

type Level zapcore.Level

const (
	LevelMin = zapcore.DebugLevel - 10
	LevelMax = zapcore.FatalLevel
)

func (l Level) Verbosity() int {
	if l > Level(zapcore.DebugLevel) {
		return 0
	}

	return -int(l) - 1
}

func (l *Level) UnmarshalText(text []byte) error {
	if bytes.HasPrefix(text, []byte("debug")) {
		vs := string(text[5:])
		var v int
		if len(vs) > 0 {
			v, _ = strconv.Atoi(vs)
		}

		*l = Level(zapcore.DebugLevel - zapcore.Level(v))

		return nil
	}

	ll, err := zapcore.ParseLevel(string(text))
	if err != nil {
		return err
	}

	*l = Level(ll)
	return nil
}

func (l Level) MarshalText() ([]byte, error) {
	return []byte(l.String()), nil
}

func (l Level) String() string {
	if l < Level(zapcore.DebugLevel) {
		return zapcore.DebugLevel.String() + strconv.Itoa(l.Verbosity())
	}

	return zapcore.Level(l).String()
}
