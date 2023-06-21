// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net_test

import (
	"crypto/rand"
	"net"

	netx "github.com/stv0g/cunicu/pkg/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("PacketPipeConn", func() {
	var l1, l2 net.Addr
	var ppc1, ppc2 net.PacketConn
	var buf1, buf2 []byte

	BeforeEach(func() {
		l1 = &net.UDPAddr{
			IP:   net.ParseIP("10.0.0.1"),
			Port: 1234,
		}

		l2 = &net.UDPAddr{
			IP:   net.ParseIP("10.0.0.2"),
			Port: 1234,
		}

		ppc1, ppc2 = netx.NewPacketPipeConn(l1, l2, 128)

		// Prepare test data
		buf1 = make([]byte, 100)
		n, err := rand.Read(buf1)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(100))
		Expect(buf1).To(HaveLen(n))

		buf2 = make([]byte, 100)
	})

	AfterEach(func() {
		err := ppc1.Close()
		Expect(err).To(Succeed())

		err = ppc2.Close()
		Expect(err).To(Succeed())
	})

	sendReceive := func(wr, rd net.PacketConn) {
		n, err := wr.WriteTo(buf1, rd.LocalAddr())
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))

		n, ra, err := rd.ReadFrom(buf2)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))
		Expect(ra).To(Equal(wr.LocalAddr()))
	}

	It("write to ppc1, read from ppc1", func() {
		sendReceive(ppc1, ppc2)
	})

	It("write to ppc2, read from ppc1", func() {
		sendReceive(ppc2, ppc1)
	})
})
