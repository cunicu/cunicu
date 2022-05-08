package cases_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/internal/util"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Case Suite")
}

var logger = test.SetupLogging()

var _ = BeforeSuite(func() {
	if !util.HasCapabilities(cap.NET_ADMIN) {
		Skip("Insufficient privileges")
	}
})
