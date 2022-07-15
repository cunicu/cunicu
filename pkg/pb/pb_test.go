package pb_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/test"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Protobuf Suite")
}

var _ = test.SetupLogging()
