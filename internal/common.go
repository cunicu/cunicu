package internal

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	glog "github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
	"k8s.io/klog/v2"
	"riasc.eu/wice/pkg/socket"
)

func SetupLogging() *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.99")
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// Redirect Kubernetes log to Zap
	klogger := logger.Named("k8s")
	klog.SetLogger(zapr.NewLogger(klogger))

	// Redirect libp2p / ipfs log to Zap
	glog.SetPrimaryCore(logger.Core())

	// Redirect gRPC log to Zap
	glogger := logger.Named("grpc")
	grpclog.SetLoggerV2(socket.NewLogger(glogger, 0))

	zap.ReplaceGlobals(logger)

	return logger
}

func SetupRand() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func SetupSignals() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	return ch
}
