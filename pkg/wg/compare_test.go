package wg_test

import (
	"github.com/stv0g/cunicu/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("peer compare", func() {

	When("equal", func() {
		var a wgtypes.Peer

		BeforeEach(func() {
			a = wgtypes.Peer{}
		})

		It("works", func() {
			Expect(wg.CmpPeers(&a, &a)).To(BeZero())
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
			Expect(wg.CmpPeers(&a, &b)).NotTo(BeZero())
		})
	})
})
