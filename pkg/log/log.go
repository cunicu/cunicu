// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package log implements adapters between logging types of various used packages
package log

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"

	"cunicu.li/cunicu/pkg/tty"
)

//nolint:gochecknoglobals
var (
	Global *Logger
)

func openSink(path string) zapcore.WriteSyncer {
	switch path {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	case "ginkgo":
		return &ginkgoSyncWriter{ginkgo.GinkgoWriter}
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		panic(err)
	}

	return tty.NewANSIStripperSynced(f)
}

type alwaysEnabled struct{}

func (e *alwaysEnabled) Enabled(zapcore.Level) bool { return true }

func SetupLogging(rule *Filter, paths []string, color bool) (logger *Logger, err error) {
	cfg := encoderConfig{
		Time:             true,
		Level:            true,
		Name:             true,
		Message:          true,
		ConsoleSeparator: " ",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeTime:       zapcore.TimeEncoderOfLayout("15:04:05.000000"),
		EncodeDuration:   zapcore.StringDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		EncodeLevel:      levelEncoder,
	}

	if color {
		cfg.ColorTime = ColorTime
		cfg.ColorContext = ColorContext
		cfg.ColorStacktrace = ColorStacktrace
		cfg.ColorName = ColorName
		cfg.ColorCaller = ColorCaller
		cfg.ColorLevel = ColorLevel
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

	if rule != nil {
		filter.Store(rule)
	}

	zlogger := zap.New(core,
		zap.ErrorOutput(ws),
		zap.WrapCore(func(c zapcore.Core) zapcore.Core {
			return NewFilteredCore(c, &filter)
		}))

	zlogger.Level()

	zap.RedirectStdLog(zlogger)
	zap.ReplaceGlobals(zlogger)

	logger = &Logger{zlogger}

	Global = logger

	// Redirect gRPC log to Zap
	glogger := NewGRPCLogger(logger.Named("grpc"))
	grpclog.SetLoggerV2(glogger)

	return logger, nil
}
