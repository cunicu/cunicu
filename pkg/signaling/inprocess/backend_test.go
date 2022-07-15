package inprocess_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/test"
	_ "riasc.eu/wice/pkg/signaling/inprocess"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "In-Process Backend Suite")
}

var _ = test.SetupLogging()

var _ = Specify("inprocess backend", func() {
	test.RunBackendTest("inprocess", 10)
})
