package core_test

import (
	"math/rand"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"kernel.org/pub/linux/libs/security/libcap/cap"

	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
)

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Core Suite")
}

var _ = test.SetupLogging()

var _ = BeforeSuite(func() {
	if !util.HasCapabilities(cap.NET_ADMIN) {
		Skip("Insufficient privileges")
	}
})
