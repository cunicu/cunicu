package e2e_test

import (
	"flag"
	"math/rand"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"
	"riasc.eu/wice/pkg/util"

	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
)

var (
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
	rand.Seed(GinkgoRandomSeed() + int64(GinkgoParallelProcess()))

	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Suite", types.ReporterConfig{
		SlowSpecThreshold: 1 * time.Minute,
	})
}

var _ = BeforeSuite(func() {
	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	DeferCleanup(gexec.CleanupBuildArtifacts)
})

var _ = ReportAfterSuite("Write report", func(r Report) {
	r.SpecReports = nil
	reporters.GenerateJSONReport(r, "logs/report.json")
})

func SpecName() []string {
	sr := CurrentSpecReport()

	normalize := func(s string) ([]string, bool) {
		p := strings.SplitN(s, ":", 2)
		if len(p) != 2 {
			return []string{}, false
		}

		ps := strings.Split(strings.ToLower(p[0]), " ")

		return ps, true
	}

	sn := []string{}
	for _, txt := range sr.ContainerHierarchyTexts {
		if n, ok := normalize(txt); ok {
			sn = append(sn, n...)
		}
	}

	if n, ok := normalize(sr.LeafNodeText); ok {
		sn = append(sn, n...)
	}

	return sn
}
