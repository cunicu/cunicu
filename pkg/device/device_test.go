package device_test

import (
	"math/rand"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
)

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Device Suite")
}

var _ = test.SetupLogging()

var _ = BeforeSuite(func() {
	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}
})
