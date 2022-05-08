package wg_test

import (
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/wg"

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

			k, _ := wgtypes.GenerateKey()
			b.PublicKey = k
		})

		It("works", func() {
			Expect(wg.CmpPeers(&a, &b)).NotTo(BeZero())
		})
	})
})
