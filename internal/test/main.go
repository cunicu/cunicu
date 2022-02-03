package test

import (
	"os"
	"testing"

	"go.uber.org/zap/zapcore"
	"riasc.eu/wice/internal"
)

func Main(m *testing.M) {
	internal.SetupRand()
	logger := internal.SetupLogging(zapcore.DebugLevel, "logs/test.log")
	defer logger.Sync()

	os.Exit(m.Run())
}
