// Package log implements adapters between logging types of various used packages
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
	"k8s.io/klog/v2"
)

var (
	Verbosity VerbosityLevel
	Severity  zap.AtomicLevel
)

func SetupLogging(severity zapcore.Level, verbosity int, outputPaths []string, errOutputPaths []string, color bool) *zap.Logger {
	Severity = zap.NewAtomicLevelAt(severity)
	Verbosity = NewVerbosityLevelAt(verbosity)

	cfg := zap.NewDevelopmentConfig()

	cfg.Level = Severity
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.009")
	if color {
		cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	} else {
		cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.OutputPaths = outputPaths
	cfg.ErrorOutputPaths = errOutputPaths

	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	if len(cfg.ErrorOutputPaths) == 0 {
		cfg.ErrorOutputPaths = []string{"stderr"}
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	// Redirect Kubernetes log to Zap
	klogger := logger.Named("backend.k8s")
	klog.SetLogger(NewK8SLogger(klogger))

	// Redirect gRPC log to Zap
	glogger := logger.Named("grpc")
	grpclog.SetLoggerV2(NewGRPCLogger(glogger, verbosity))

	zap.RedirectStdLog(logger)
	zap.ReplaceGlobals(logger)

	return logger
}

func WithVerbose(verbose int) zap.Option {
	return zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &verbosityCore{
			Core:    core,
			verbose: verbose,
		}
	})
}

type verbosityCore struct {
	zapcore.Core
	verbose int
}

func (c *verbosityCore) Enabled(lvl zapcore.Level) bool {
	return c.Core.Enabled(lvl) && (lvl != zap.DebugLevel || Verbosity.Enabled(c.verbose))
}

func (c *verbosityCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}
