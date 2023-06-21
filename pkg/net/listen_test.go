// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net_test

import (
	"net"

	"github.com/stv0g/cunicu/pkg/config"
	netx "github.com/stv0g/cunicu/pkg/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("find bindable port in range", func() {
	Describe("next", func() {
		It("finds the next available port", func() {
			port, err := netx.FindNextPortToListen("udp", 10032, config.EphemeralPortMax)
			Expect(err).To(Succeed())

			Expect(port).To(BeNumerically("==", 10032))
		})

		It("fails if max < min", func() {
			_, err := netx.FindNextPortToListen("udp", 10010, 1005)
			Expect(err).To(MatchError("minimal port must be larger than maximal port number"))
		})

		It("fails for unsupported network type", func() {
			_, err := netx.FindNextPortToListen("tcp", 10010, 10020)
			Expect(err).To(MatchError("unsupported network: tcp"))
		})

		Context("with used port", func() {
			var conn net.Conn
			port := 10024

			BeforeEach(func() {
				var err error

				conn, err = net.ListenUDP("udp", &net.UDPAddr{
					Port: port,
				})
				Expect(err).To(Succeed())
			})

			AfterEach(func() {
				err := conn.Close()
				Expect(err).To(Succeed())
			})

			It("finds the next available port", func() {
				portFound, err := netx.FindNextPortToListen("udp", port, config.EphemeralPortMax)
				Expect(err).To(Succeed())
				Expect(portFound).To(BeNumerically("==", port+1))
			})

			It("fails with a single port if that one is already used", func() {
				_, err := netx.FindNextPortToListen("udp", port, port)
				Expect(err).To(MatchError("failed to find port"))
			})
		})
	})

	Describe("random", func() {
		It("finds a single port", func() {
			port, err := netx.FindRandomPortToListen("udp", config.EphemeralPortMin, config.EphemeralPortMax)
			Expect(err).To(Succeed())

			Expect(port).To(And(
				BeNumerically(">=", config.EphemeralPortMin),
				BeNumerically("<=", config.EphemeralPortMax),
			))
		})

		Context("finds multiple ports without collisions", func() {
			var conns []net.Conn
			cnt := 10

			BeforeEach(func() {
				conns = []net.Conn{}
			})

			AfterEach(func() {
				for _, conn := range conns {
					Expect(conn.Close()).To(Succeed())
				}
			})

			It("can create conns", func() {
				for i := 0; i < cnt; i++ {
					port, err := netx.FindRandomPortToListen("udp", 12000, 12000+cnt)
					Expect(err).To(Succeed())

					conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
					Expect(err).To(Succeed())

					conns = append(conns, conn)

					Expect(port).To(And(
						BeNumerically(">=", 12000),
						BeNumerically("<=", 12100),
					))
				}
			})
		})

		It("fails if max < min", func() {
			_, err := netx.FindRandomPortToListen("udp", 10010, 1005)
			Expect(err).To(MatchError("minimal port must be larger than maximal port number"))
		})

		It("fails for unsupported network type", func() {
			_, err := netx.FindRandomPortToListen("tcp", 10010, 10020)
			Expect(err).To(MatchError("unsupported network: tcp"))
		})

		It("works with a single port", func() {
			port, err := netx.FindRandomPortToListen("udp", 10012, 10012)
			Expect(err).To(Succeed())

			Expect(port).To(BeNumerically("==", 10012))
		})

		Context("with used port", func() {
			var conn net.Conn
			port := 10011

			BeforeEach(func() {
				var err error

				conn, err = net.ListenUDP("udp", &net.UDPAddr{
					Port: port,
				})
				Expect(err).To(Succeed())
			})

			AfterEach(func() {
				err := conn.Close()
				Expect(err).To(Succeed())
			})

			It("fails with a single port if that one is already used", func() {
				_, err := netx.FindRandomPortToListen("udp", port, port)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
