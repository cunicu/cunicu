package test

import (
	"net/url"

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
	zap.RegisterSink("ginkgo", func(u *url.URL) (zap.Sink, error) {
		return &writerWrapper{ginkgo.GinkgoWriter}, nil
	})

	return log.SetupLogging(zap.DebugLevel, []string{"ginkgo:"}, []string{"ginkgo:"})
}
