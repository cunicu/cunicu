package grpc_test

import (
	"net"
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/pkg/signaling/grpc"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gRPC Backend Suite")
}

var _ = Describe("gRPC backend", func() {
	var svr *grpc.Server
	var l *net.TCPListener

	BeforeEach(func() {
		var err error
		l, err = net.ListenTCP("tcp", &net.TCPAddr{
			IP: net.IPv6loopback,
		})
		Expect(err).To(Succeed(), "Failed to listen: %s", err)

		// Start local dummy gRPC server
		svr = grpc.NewServer()
		go svr.Serve(l)
	})

	It("works", func() {
		u := url.URL{
			Scheme:   "grpc",
			Host:     l.Addr().String(),
			RawQuery: "insecure=true",
		}
		test.RunBackendTest(u.String(), 10)
	})

	AfterEach(func() {
		svr.Stop()
	})
})
