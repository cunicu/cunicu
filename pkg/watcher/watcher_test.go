package watcher_test

import (
	"math/rand"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/test"
)

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Watcher Suite")
}

var _ = test.SetupLogging()

var _ = BeforeSuite(func() {
	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}
})
