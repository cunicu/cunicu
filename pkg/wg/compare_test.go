// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg_test

import (
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("compare", func() {
	Context("peer", func() {
		When("equal", func() {
			var a wgtypes.Peer

			BeforeEach(func() {
				a = wgtypes.Peer{}
			})

			It("works", func() {
				Expect(wg.CmpPeers(a, a)).To(BeZero())
			})
		})

		When("unequal", func() {
			var a, b wgtypes.Peer

			BeforeEach(func() {
				a = wgtypes.Peer{}
				b = wgtypes.Peer{}

				k, err := wgtypes.GenerateKey()
				Expect(err).To(Succeed())

				b.PublicKey = k
			})

			It("works", func() {
				Expect(wg.CmpPeers(a, b)).NotTo(BeZero())
			})
		})
	})

	Context("device", func() {
		var err error
		var sk1, sk2 wgtypes.Key

		BeforeEach(func() {
			sk1, err = wgtypes.GeneratePrivateKey()
			Expect(err).To(Succeed())

			sk2, err = wgtypes.GeneratePrivateKey()
			Expect(err).To(Succeed())
		})

		It("equal", func() {
			d1 := wgtypes.Device{
				PublicKey: sk1.PublicKey(),
			}

			Expect(wg.CmpDevices(d1, d1)).To(Equal(0))
		})

		It("unequal", func() {
			d1 := wgtypes.Device{
				PublicKey: sk1.PublicKey(),
			}

			d2 := wgtypes.Device{
				PublicKey: sk2.PublicKey(),
			}

			Expect(wg.CmpDevices(d1, d2)).NotTo(Equal(0))
		})
	})

	Context("last handshake time", func() {
		var p1, p2 wgtypes.Peer
		var now time.Time

		BeforeEach(func() {
			now = time.Now()

			p1 = wgtypes.Peer{
				LastHandshakeTime: now,
			}

			p2 = wgtypes.Peer{
				LastHandshakeTime: now.Add(time.Second),
			}
		})

		It("equal", func() {
			Expect(wg.CmpPeerHandshakeTime(p1, p1)).To(Equal(0))
		})

		It("before", func() {
			Expect(wg.CmpPeerHandshakeTime(p1, p2)).To(BeNumerically(">", 0))
		})

		It("after", func() {
			Expect(wg.CmpPeerHandshakeTime(p2, p1)).To(BeNumerically("<", 0))
		})
	})
})
