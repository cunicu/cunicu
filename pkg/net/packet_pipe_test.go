// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net_test

import (
	"net"
	"os"
	"time"

	netx "github.com/stv0g/cunicu/pkg/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("packet pipe", func() {
	var pp netx.PacketPipe

	lAddr := &net.UDPAddr{
		IP:   net.ParseIP("1.2.3.4"),
		Port: 1234,
	}

	rAddr := &net.UDPAddr{
		IP:   net.ParseIP("1.2.3.4"),
		Port: 1234,
	}

	BeforeEach(func() {
		pp = *netx.NewPacketPipe(lAddr, 0)
	})

	AfterEach(func() {
		err := pp.Close()
		Expect(err).To(Succeed())
	})

	It("has correct local address", func() {
		Expect(pp.LocalAddr()).To(Equal(lAddr))
	})

	It("read deadline", func() {
		buf := []byte{}

		err := pp.SetReadDeadline(time.Now().Add(time.Second))
		Expect(err).To(Succeed())

		_, _, err = pp.ReadFrom(buf)
		Expect(err).To(MatchError(os.ErrDeadlineExceeded))
	})

	It("write deadline", func() {
		buf := []byte{}

		err := pp.SetWriteDeadline(time.Now().Add(time.Second))
		Expect(err).To(Succeed())

		_, err = pp.WriteFrom(buf, lAddr)
		Expect(err).To(MatchError(os.ErrDeadlineExceeded))
	})

	It("can receive to oversized buffer", func() {
		buf1 := []byte("1234567890")
		n1 := make(chan int)
		err1 := make(chan error)

		go func() {
			n, err := pp.WriteFrom(buf1, rAddr)
			n1 <- n
			err1 <- err
		}()

		buf2 := make([]byte, 20)

		n2, rAddr2, err := pp.ReadFrom(buf2)

		Expect(err).To(Succeed())
		Expect(n2).To(Equal(10))
		Expect(rAddr2).To(Equal(rAddr))
		Expect(buf2[:n2]).To(Equal(buf1[:n2]))

		Eventually(n1).Should(Receive(Equal(10)))
		Eventually(err1).Should(Receive(Succeed()))
	})

	It("can receive to undersized buffer", func() {
		buf1 := []byte("1234567890")
		n1 := make(chan int)
		err1 := make(chan error)

		go func() {
			n, err := pp.WriteFrom(buf1, rAddr)
			n1 <- n
			err1 <- err
		}()

		buf2 := make([]byte, 5)

		n2, rAddr2, err := pp.ReadFrom(buf2)

		Expect(err).To(Succeed())
		Expect(n2).To(Equal(5))
		Expect(rAddr2).To(Equal(rAddr))
		Expect(buf2[:n2]).To(Equal(buf1[:n2]))

		Eventually(n1).Should(Receive(Equal(10)))
		Eventually(err1).Should(Receive(Succeed()))
	})
})
