// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net_test

import (
	"net"
	"os"
	"time"

	"github.com/pion/stun"

	"github.com/stv0g/cunicu/pkg/log"
	netx "github.com/stv0g/cunicu/pkg/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("STUNPacketHandler", func() {
	var l1, l2 net.Addr
	var ppc1, ppc2 net.PacketConn
	var fc *netx.FilteredConn
	var buf1, buf2, buf3 []byte

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

		fc = netx.NewFilteredConn(ppc2, log.Global)
	})

	AfterEach(func() {
		err := ppc1.Close()
		Expect(err).To(Succeed())

		err = ppc2.Close()
		Expect(err).To(Succeed())

		err = fc.Close()
		Expect(err).To(Succeed())
	})

	It("can handle STUN", func() {
		fc.AddPacketReadHandler(&netx.STUNPacketHandler{
			Logger: log.Global,
		})

		m := stun.New()
		m.Encode()

		buf1 = m.Raw
		n, err := ppc1.WriteTo(buf1, nil)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))

		err = ppc2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		Expect(err).To(Succeed())

		buf2 = make([]byte, len(buf1))
		_, _, err = fc.ReadFrom(buf2)
		Expect(err).To(MatchError(os.ErrDeadlineExceeded))
	})

	It("can handle STUN with conn", func() {
		stunConn := fc.AddPacketReadHandlerConn(&netx.STUNPacketHandler{
			Logger: log.Global,
		})

		m := stun.New()
		m.Encode()

		buf1 = m.Raw
		n, err := ppc1.WriteTo(buf1, l2)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))

		err = ppc2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		Expect(err).To(Succeed())

		buf2 = make([]byte, len(buf1))
		_, _, err = fc.ReadFrom(buf2)
		Expect(err).To(MatchError(os.ErrDeadlineExceeded))

		buf3 = make([]byte, 100)
		n, ra, err := stunConn.ReadFrom(buf3)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(m.Raw)))
		Expect(ra).To(Equal(ppc1.LocalAddr()))
	})
})
