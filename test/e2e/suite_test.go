// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/onsi/ginkgo/v2/reporters"

	osx "github.com/stv0g/cunicu/pkg/os"
	"github.com/stv0g/cunicu/test/e2e/nodes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//nolint:gochecknoglobals
var options testOptions

type testOptions struct {
	setup   bool
	persist bool
	capture bool
	debug   bool
	timeout time.Duration
}

// Register your flags in an init function.  This ensures they are registered _before_ `go test` calls flag.Parse().
func init() { //nolint:gochecknoinits
	flag.BoolVar(&options.setup, "setup", false, "Do not run the actual tests, but stop after test-network setup")
	flag.BoolVar(&options.persist, "persist", false, "Do not tear-down virtual network")
	flag.BoolVar(&options.capture, "capture", false, "Captures network-traffic to PCAPng file")
	flag.BoolVar(&options.debug, "debug", false, "Start debugging agents and signaling servers")
	flag.DurationVar(&options.timeout, "timeout", 10*time.Minute, "Timeout for connectivity tests")
}

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Suite")
}

var _ = BeforeSuite(func() {
	if !osx.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}

	if options.setup && !options.persist {
		GinkgoT().Log("Persisting Gont network as --setup was requested")
		options.persist = true
	}

	DeferCleanup(nodes.CleanupBinary)
})

var _ = ReportAfterSuite("Write report", func(r Report) {
	r.SpecReports = nil

	if err := os.MkdirAll("logs", 0o755); err != nil {
		panic(err)
	}

	if err := reporters.GenerateJSONReport(r, "logs/report.json"); err != nil {
		panic(err)
	}
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
