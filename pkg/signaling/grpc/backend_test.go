// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc_test

import (
	"net"
	"net/url"
	"testing"

	"github.com/stv0g/cunicu/pkg/signaling/grpc"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "gRPC Backend Suite")
}

var _ = Describe("gRPC backend", func() {
	var svr *grpc.Server
	var l *net.TCPListener
	var u url.URL

	BeforeEach(func() {
		var err error
		l, err = net.ListenTCP("tcp", &net.TCPAddr{
			IP: net.IPv6loopback,
		})
		Expect(err).To(Succeed(), "Failed to listen: %s", err)

		// Start local dummy gRPC server
		svr = grpc.NewSignalingServer()
		go svr.Serve(l) //nolint:errcheck

		u = url.URL{
			Scheme:   "grpc",
			Host:     l.Addr().String(),
			RawQuery: "insecure=true",
		}
	})

	test.BackendTest(&u, 10)

	AfterEach(func() {
		svr.Stop()
	})
})
