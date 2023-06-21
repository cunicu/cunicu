// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net_test

import (
	"net"

	netx "github.com/stv0g/cunicu/pkg/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("split conn", func() {
	var err error
	var split, send, recv net.PacketConn
	var buf1 []byte

	BeforeEach(func() {
		buf1 = []byte("1234567890")
		loAddr := &net.UDPAddr{IP: net.IPv6loopback}

		send, err = net.ListenUDP("udp", loAddr)
		Expect(err).To(Succeed())

		recv, err = net.ListenUDP("udp", loAddr)
		Expect(err).To(Succeed())

		split = netx.NewSplitConn(recv, send)
	})

	It("works", func() {
		n1, err := split.WriteTo(buf1, recv.LocalAddr())
		Expect(err).To(Succeed())
		Expect(n1).To(Equal(10))

		buf2 := make([]byte, 20)
		n2, rAddr, err := split.ReadFrom(buf2)
		Expect(err).To(Succeed())
		Expect(n2).To(Equal(10))

		Expect(buf1).To(Equal(buf2[:n2]))
		Expect(rAddr).To(Equal(send.LocalAddr()))
	})

	AfterEach(func() {
		err = split.Close()
		Expect(err).To(Succeed())
	})
})
