package main_test

import (
	"os"
	"testing"

	"github.com/go-logr/zapr"
	glog "github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
	"k8s.io/klog/v2"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/log"
)

func setupLogging() *zap.Logger {
	cfg := zap.NewDevelopmentEncoderConfig()
	cfg.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.99")

	consoleEncoder := zapcore.NewConsoleEncoder(cfg)

	if err := os.MkdirAll("./logs", 0755); err != nil {
		panic("failed to create log dir")
	}

	f, err := os.OpenFile("logs/test.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic("failed to open log file")
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(f), zap.InfoLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)

	logger := zap.New(core)

	zap.RedirectStdLog(logger)
	zap.ReplaceGlobals(logger)
	zap.LevelFlag("log-level", zap.DebugLevel, "Log level")

	// Redirect Kubernetes log to Zap
	klogger := logger.Named("k8s")
	klog.SetLogger(zapr.NewLogger(klogger))

	// Redirect libp2p / ipfs log to Zap
	glog.SetPrimaryCore(logger.Core())

	// Redirect gRPC log to Zap
	glogger := logger.Named("grpc")
	grpclog.SetLoggerV2(log.NewGRPCLogger(glogger))

	zap.ReplaceGlobals(logger)

	return logger
}

func TestMain(m *testing.M) {
	internal.SetupRand()
	logger := setupLogging()
	defer logger.Sync()

	os.Exit(m.Run())
}
