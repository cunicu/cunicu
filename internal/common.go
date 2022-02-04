package internal

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	glog "github.com/ipfs/go-log/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
	"k8s.io/klog/v2"
	"riasc.eu/wice/internal/log"
)

func SetupLogging(level zapcore.Level, file string) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()

	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.99")
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true

	if file != "" {
		cfg.OutputPaths = append(cfg.OutputPaths, file)
		path := filepath.Dir(file)
		if err := os.MkdirAll(path, 0755); err != nil {
			panic("failed to create log directory: " + err.Error())
		}
	}

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
	grpclog.SetLoggerV2(log.NewGRPCLogger(glogger))

	zap.RedirectStdLog(logger)
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

func SetupPeriodicHeapDumps() {
	logger := zap.L().Named("pprof")

	hn, _ := os.Hostname()
	s := strings.Split(hn, ".")
	prefix := s[0]

	go func() {
		for i := 0; ; i++ {
			fn := fmt.Sprintf("%s_heap.dump", prefix)

			wr, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Error("Failed to open file for heap profile", zap.Error(err))
				continue
			}

			if err := pprof.WriteHeapProfile(wr); err != nil {
				logger.Error("Failed to write heap profile", zap.Error(err))
			}

			logger.Debug("Wrote heap dump", zap.String("file", fn))

			time.Sleep(1 * time.Second)
		}
	}()
}
