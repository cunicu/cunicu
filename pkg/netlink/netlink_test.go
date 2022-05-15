package netlink_test

import (
	"fmt"
	"math/rand"
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
	var link *nl.Wireguard
	var linkName string

	BeforeAll(func() {
		linkName = fmt.Sprintf("wg-test-%d", rand.Intn(1000))

		if !util.HasCapabilities(cap.NET_ADMIN) {
			Skip("Insufficient privileges")
		}

		link = &nl.Wireguard{
			LinkAttrs: netlink.NewLinkAttrs(),
		}
		link.LinkAttrs.Name = linkName
	})

	It("can add a link", func() {
		Expect(netlink.LinkAdd(link)).To(Succeed())
	})

	It("can get link by name", func() {
		l2, err := netlink.LinkByName(linkName)

		Expect(err).To(Succeed())
		Expect(l2.Type()).To(Equal("wireguard"))
	})

	It("can delete link again", func() {
		Expect(netlink.LinkDel(link)).To(Succeed())
	})
})
