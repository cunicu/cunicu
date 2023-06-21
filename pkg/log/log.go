// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package log implements adapters between logging types of various used packages
package log

import (
	"os"

	"github.com/stv0g/cunicu/pkg/tty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
)

//nolint:gochecknoglobals
var (
	Rule   AtomicFilterRule
	Global *Logger
)

func DebugLevel(verbosity int) Level {
	return Level(zapcore.DebugLevel) - Level(verbosity)
}

func openSink(path string) zapcore.WriteSyncer {
	if path == "stdout" {
		return os.Stdout
	} else if path == "stderr" {
		return os.Stderr
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}

	return tty.NewANSIStripperSynced(f)
}

type alwaysEnabled struct{}

func (e *alwaysEnabled) Enabled(zapcore.Level) bool { return true }

func SetupLogging(rule string, paths []string, color bool) (logger *Logger, err error) {
	cfg := encoderConfig{
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:          "T",
			LevelKey:         "L",
			NameKey:          "N",
			CallerKey:        "C",
			FunctionKey:      zapcore.OmitKey,
			MessageKey:       "M",
			StacktraceKey:    "S",
			ConsoleSeparator: " ",
			LineEnding:       zapcore.DefaultLineEnding,
			EncodeTime:       zapcore.TimeEncoderOfLayout("15:04:05.000000"),
			EncodeDuration:   zapcore.StringDurationEncoder,
			EncodeCaller:     zapcore.ShortCallerEncoder,
			EncodeLevel:      levelEncoder,
		},
	}

	if color {
		cfg.ColorTime = ColorTime
		cfg.ColorContext = ColorContext
		cfg.ColorStacktrace = ColorStacktrace
		cfg.ColorName = ColorName
		cfg.ColorCaller = ColorCaller
		cfg.ColorLevel = ColorLevel
	} else {
		cfg.ColorLevel = func(lvl zapcore.Level) string {
			return ""
		}
	}

	wss := []zapcore.WriteSyncer{}

	for _, path := range paths {
		wss = append(wss, openSink(path))
	}

	if len(wss) == 0 {
		wss = append(wss, os.Stdout)
	}

	ws := zapcore.NewMultiWriteSyncer(wss...)
	enc := newEncoder(cfg)
	core := zapcore.NewCore(enc, ws, &alwaysEnabled{})

	if rule == "" {
		rule = "*"
	}

	filterRule, err := ParseFilterRule(rule)
	if err != nil {
		return nil, err
	}

	Rule.Store(filterRule)

	zlogger := zap.New(core,
		zap.ErrorOutput(ws),
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return NewFilteredCore(c, &Rule)
		}))

	zlogger.Level()

	zap.RedirectStdLog(zlogger)
	zap.ReplaceGlobals(zlogger)

	logger = &Logger{zlogger}

	Global = logger

	// Redirect gRPC log to Zap
	glogger := logger.Named("grpc")
	grpclog.SetLoggerV2(NewGRPCLogger(glogger, Level(zlogger.Level()).Verbosity()))

	return logger, nil
}

type forceReflect struct {
	any
}

func ForceReflect(key string, val any) zapcore.Field {
	return zapcore.Field{Key: key, Type: zapcore.ReflectType, Interface: forceReflect{val}}
}
