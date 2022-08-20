package test_test

import (
	"flag"
	"math/rand"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/util"
)

var (
	logger *zap.Logger

	setup   bool
	persist bool
	capture bool
)

// Register your flags in an init function.  This ensures they are registered _before_ `go test` calls flag.Parse().
func init() {
	flag.BoolVar(&setup, "setup", false, "Do not run the actual tests, but stop after test-network setup")
	flag.BoolVar(&persist, "persist", false, "Do not tear-down virtual network")
	flag.BoolVar(&capture, "capture", false, "Captures network-traffic to PCAPng file")
}

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Suite", types.ReporterConfig{
		SlowSpecThreshold: 5 * time.Minute,
	})
}

var _ = BeforeSuite(func() {
	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	DeferCleanup(gexec.CleanupBuildArtifacts)
})
