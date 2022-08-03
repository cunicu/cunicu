package test

import (
	"net/url"
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/log"
)

type writerWrapper struct {
	ginkgo.GinkgoWriterInterface
}

func (w *writerWrapper) Close() error {
	return nil
}

func (w *writerWrapper) Sync() error {
	return nil
}

func SetupLogging() *zap.Logger {
	return SetupLoggingWithFile("", false)
}

func SetupLoggingWithFile(fn string, truncate bool) *zap.Logger {
	if err := zap.RegisterSink("ginkgo", func(u *url.URL) (zap.Sink, error) {
		return &writerWrapper{ginkgo.GinkgoWriter}, nil
	}); err != nil {
		panic(err)
	}

	outputPaths := []string{"ginkgo:"}

	if fn != "" {
		// Truncate log file if requested
		if truncate {
			//#nosec G104 -- May fail if file does not exist yet
			os.Truncate(fn, 0)
		}

		// Create parent directories for log file
		if path := path.Dir(fn); path != "" {
			if err := os.MkdirAll(path, 0750); err != nil {
				panic(err)
			}
		}

		outputPaths = append(outputPaths, fn)
	}

	return log.SetupLogging(zap.DebugLevel, outputPaths, outputPaths)
}
