package netlink_test

import (
	"testing"

	"github.com/vishvananda/netlink"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"riasc.eu/wice/internal/util"
	nl "riasc.eu/wice/pkg/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Netlink Suite")
}

var _ = Describe("Wireguard link handling", Ordered, func() {
	var l *nl.Wireguard

	BeforeAll(func() {
		if !util.HasCapabilities(cap.NET_ADMIN) {
			Skip("Insufficient privileges")
		}

		l = &nl.Wireguard{
			LinkAttrs: netlink.NewLinkAttrs(),
		}
		l.LinkAttrs.Name = "wg-test0"
	})

	It("can add a link", func() {
		Expect(netlink.LinkAdd(l)).To(Succeed())
	})

	It("can get link by name", func() {
		l2, err := netlink.LinkByName("wg-test0")

		Expect(err).To(Succeed())
		Expect(l2.Type()).To(Equal("wireguard"))
	})

	It("can delete link again", func() {
		Expect(netlink.LinkDel(l)).To(Succeed())
	})
})
