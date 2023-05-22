package log_test

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pion/zapion"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/tty"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
	"k8s.io/klog/v2"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logging Suite")
}

// TODO: This test is currently broken on Windows due:
// https://github.com/uber-go/zap/issues/621
var _ = Context("log", Label("broken-on-windows"), func() {
	var logger *zap.Logger
	var lvl zapcore.Level
	var logPath, msg, scope string

	BeforeEach(func() {
		tmpDir := GinkgoT().TempDir()

		logPath = filepath.Join(tmpDir, "std.log")
		msg = fmt.Sprintf("Test message %s", tty.Mods("something red", tty.FgRed))

		os.Setenv("GRPC_GO_LOG_VERBOSITY_LEVEL", "2")
		os.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", lvl.String())
		os.Setenv("PION_LOG", lvl.String())
	})

	JustBeforeEach(func() {
		logger = log.SetupLogging(lvl, 0, []string{logPath}, nil, true)
	})

	Context("simple", func() {
		It("can log via created logger", func() {
			scope = ""
			logger.Info(msg)
		})

		It("can log via std logger", func() {
			scope = ""
			stdlog.Print(msg)
		})

		It("can log via global logger", func() {
			scope = ""
			zap.L().Info(msg)
		})

		It("can log via pion logger", func() {
			lf := zapion.ZapFactory{
				BaseLogger: logger.Named("ice"),
			}
			logger := lf.NewLogger("myscope")

			scope = "ice.myscope"
			logger.Info(msg)
		})

		It("can log via gRPC logger", func() {
			scope = "grpc"
			grpclog.Info(msg)
		})

		It("can log via k8s logger", func() {
			scope = "backend.k8s"
			klog.Info(msg)
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

		if scope != "" {
			scope += `\t`
		}

		regex := fmt.Sprintf(`\d{2}:\d{2}:\d{2}.\d{3}\t%s\t%s%s`,
			regexp.QuoteMeta(tty.Mods(lvl.String(), tty.FgBlue)), scope,
			regexp.QuoteMeta(msg))

		Expect(logContents).To(MatchRegexp(regex), "Log output '%s' does not match regex '%s'", logContents, regex)
	})
})
