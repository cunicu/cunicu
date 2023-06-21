// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net_test

import (
	"net"
	"os"
	"time"

	"github.com/stv0g/cunicu/pkg/log"
	netx "github.com/stv0g/cunicu/pkg/net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("FilteredConn", func() {
	var l1, l2 net.Addr
	var ppc1, ppc2 net.PacketConn
	var fc *netx.FilteredConn
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

	It("can filter", func() {
		fc.AddPacketReadHandler(&functionHandler{
			handler: func(buf []byte, addr net.Addr) (bool, error) {
				Expect(buf).To(Equal(buf2))
				Expect(addr).To(Equal(l1))

				return false, nil
			},
		})

		buf1 = []byte("1234567890")
		n, err := ppc1.WriteTo(buf1, nil)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))

		buf2 = make([]byte, len(buf1))
		n, ra, err := fc.ReadFrom(buf2)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))
		Expect(ra).To(Equal(ppc1.LocalAddr()))
	})

	It("can filter and drop", func() {
		fc.AddPacketReadHandler(&functionHandler{
			handler: func(buf []byte, addr net.Addr) (bool, error) {
				Expect(buf).To(Equal(buf2))
				Expect(addr).To(Equal(l1))

				// Abort on packets with an even first byte
				if len(buf) > 0 && buf[0]%2 == 0 {
					return true, nil
				}

				return false, nil
			},
		})

		buf1 = []byte("0123456789")
		n, err := ppc1.WriteTo(buf1, nil)
		Expect(err).To(Succeed())
		Expect(n).To(Equal(len(buf1)))

		err = ppc2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		Expect(err).To(Succeed())

		buf2 = make([]byte, len(buf1))
		_, _, err = fc.ReadFrom(buf2)
		Expect(err).To(MatchError(os.ErrDeadlineExceeded))
	})
})

type functionHandler struct {
	handler func(buf []byte, addr net.Addr) (bool, error)
}

func (h *functionHandler) OnPacketRead(buf []byte, addr net.Addr) (bool, error) {
	return h.handler(buf, addr)
}
