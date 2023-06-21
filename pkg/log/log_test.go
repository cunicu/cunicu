// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package log_test

import (
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"testing"

	"github.com/pion/zapion"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"

	"github.com/stv0g/cunicu/pkg/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logging Suite")
}

// TODO: This test is currently broken on Windows due:
// https://github.com/uber-go/zap/issues/621
var _ = Context("log", Label("broken-on-windows"), func() {
	var logger *log.Logger
	var lvl zapcore.Level
	var logPath, msg, name string

	BeforeEach(func() {
		tmpDir := GinkgoT().TempDir()

		logPath = filepath.Join(tmpDir, "std.log")
		msg = "Test message"

		os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "2")
		os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", lvl.String())
		os.Setenv("PION_LOG", lvl.String())

		log.ResetWidths()
	})

	JustBeforeEach(func() {
		var err error
		logger, err = log.SetupLogging("", []string{logPath}, false)
		Expect(err).To(Succeed())
	})

	Context("simple", func() {
		It("can log via created logger", func() {
			name = ""
			logger.Info(msg)
		})

		It("can log via std logger", func() {
			name = ""
			stdlog.Print(msg)
		})

		It("can log via global logger", func() {
			name = ""
			log.Global.Info(msg)
		})

		It("can log via pion logger", func() {
			lf := zapion.ZapFactory{
				BaseLogger: logger.Named("ice").Logger,
			}
			logger := lf.NewLogger("myscope")

			name = "ice.myscope"
			logger.Info(msg)
		})

		It("can log via gRPC logger", func() {
			name = "grpc"
			grpclog.Info(msg)
		})
	})

	AfterEach(func() {
		err := logger.Sync()
		Expect(err).To(Succeed(), "Failed to sync logger: %s", err)

		Expect(logPath).To(BeARegularFile(), "Standard log does not exist")

		logFile, err := os.Open(logPath)
		Expect(err).To(Succeed(), "Failed to open standard log file: %s", err)

		logContents, err := io.ReadAll(logFile)
		Expect(err).To(Succeed(), "Failed to read standard log contents: %s", err)
		Expect(logContents).NotTo(BeEmpty())

		regexTime := `\d{2}:\d{2}:\d{2}.\d{6} `
		regexLevel := lvl.String() + " "
		regexName := name + " "

		var regex string
		if name != "" {
			regex = regexTime + regexLevel + regexName + msg
		} else {
			regex = regexTime + regexLevel + msg
		}

		Expect(string(logContents)).To(MatchRegexp(regex), "Log output '%s' does not match regex '%s'", logContents, regex)
	})
})
