package test

import (
	"net/url"
	"os"
	"path"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/internal/log"
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

	if truncate {
		os.Truncate(fn, 0)
	}

	if fn != "" {
		if path := path.Dir(fn); path != "" {
			if err := os.MkdirAll(path, 0755); err != nil {
				panic(err)
			}
		}

		outputPaths = append(outputPaths, fn)
	}

	return log.SetupLogging(zap.DebugLevel, outputPaths, outputPaths)
}
