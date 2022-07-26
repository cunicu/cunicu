package cases_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Test Case Suite")
}

var _ = test.SetupLoggingWithFile("logs/test.log", true)

var _ = BeforeSuite(func() {
	if !util.HasCapabilities(cap.NET_ADMIN) {
		Skip("Insufficient privileges")
	}
})
