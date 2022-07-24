package inprocess_test

import (
	"net/url"
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

var _ = Describe("inprocess backend", func() {
	u := url.URL{
		Scheme: "inprocess",
	}

	test.BackendTest(&u, 10)
})
