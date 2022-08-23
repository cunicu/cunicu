package test

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/log"
	"riasc.eu/wice/pkg/util"
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
		return &writerWrapper{
			GinkgoWriterInterface: ginkgo.GinkgoWriter,
		}, nil
	}); err != nil && !strings.Contains(err.Error(), "already registered") {
		panic(err)
	}

	outputPaths := []string{"ginkgo:"}

	if fn != "" {
		// Create parent directories for log file
		if path := path.Dir(fn); path != "" {
			if err := os.MkdirAll(path, 0750); err != nil {
				panic(fmt.Errorf("failed to directory of log file: %w", err))
			}
		}

		fl := os.O_CREATE | os.O_APPEND | os.O_WRONLY
		if truncate {
			fl |= os.O_TRUNC
		}

		f, err := os.OpenFile(fn, fl, 0644)
		if err != nil {
			panic(fmt.Errorf("failed to open log file '%s': %w", fn, err))
		}

		ginkgo.GinkgoWriter.TeeTo(&util.ANSIStripper{
			Writer: f,
		})
	}

	return log.SetupLogging(zap.DebugLevel, outputPaths, outputPaths)
}
