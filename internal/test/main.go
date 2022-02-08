package test

import (
	"os"
	"testing"

	"go.uber.org/zap/zapcore"
	"riasc.eu/wice/internal"
)

func Main(m *testing.M) {
	internal.SetupRand()

	logPath := "logs/test.log"

	if err := os.RemoveAll(logPath); err != nil {
		panic(err)
	}

	logger := internal.SetupLogging(zapcore.DebugLevel, logPath)
	defer logger.Sync()

	os.Exit(m.Run())
}
