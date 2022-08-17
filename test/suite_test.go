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
	"riasc.eu/wice/pkg/util"
)

var (
	persist    bool
	capture    bool
	binaryPath string
)

// Register your flags in an init function.  This ensures they are registered _before_ `go test` calls flag.Parse().
func init() {
	flag.BoolVar(&persist, "persist", false, "Do not tear-down virtual network")
	flag.BoolVar(&capture, "capture", false, "Captures network-traffic to PCAPng file")
	flag.StringVar(&binaryPath, "binary", "", "Filename of wice binary (Default: build during test execution)")
}

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Suite", types.ReporterConfig{
		SlowSpecThreshold: 5 * time.Minute,
	})
}

var _ = BeforeSuite(func() {
	var err error

	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	if binaryPath == "" {
		binaryPath, err = gexec.Build("../cmd")
		Expect(err).NotTo(HaveOccurred())
	}

	DeferCleanup(gexec.CleanupBuildArtifacts)
})
