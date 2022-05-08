package p2p_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/test"
	_ "riasc.eu/wice/pkg/signaling/inprocess"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LibP2P Backend Suite")
}

var _ = Specify("libp2p backend", func() {
	test.RunBackendTest("p2p:?private=true&mdns=true", 2)
})
