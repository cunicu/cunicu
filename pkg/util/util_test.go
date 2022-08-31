package util_test

import (
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/test"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utilities Suite")
}

var _ = test.SetupLogging()

var _ = Describe("utils", func() {

	Context("endpoint comparisons", func() {
		It("to be equal", func() {
			a := net.UDPAddr{
				IP:   net.ParseIP("1.1.1.1"),
				Port: 1,
			}

			Expect(util.CmpEndpoint(&a, &a)).To(BeZero())
		})

		It("to be unequal", func() {
			a := net.UDPAddr{
				IP:   net.ParseIP("1.1.1.1"),
				Port: 1,
			}

			b := net.UDPAddr{
				IP:   net.ParseIP("2.2.2.2"),
				Port: 1,
			}

			Expect(util.CmpEndpoint(&a, &b)).NotTo(BeZero())
		})

		It("nil to be equal", func() {
			Expect(util.CmpEndpoint(nil, nil)).To(BeZero())
		})

		It("mixed nil to be unequal", func() {
			a := net.UDPAddr{
				IP:   net.ParseIP("1.1.1.1"),
				Port: 1,
			}

			Expect(util.CmpEndpoint(&a, nil)).NotTo(BeZero())
			Expect(util.CmpEndpoint(nil, &a)).NotTo(BeZero())
		})
	})

	Context("network comparisons", func() {
		It("compare equal networks", func() {
			_, a, err := net.ParseCIDR("1.1.1.1/0")
			Expect(err).To(Succeed())

			Expect(util.CmpNet(a, a)).To(BeZero())
		})

		It("compare unequal networks", func() {
			_, a, err := net.ParseCIDR("1.1.1.1/0")
			Expect(err).To(Succeed())

			_, b, err := net.ParseCIDR("1.1.1.1/1")
			Expect(err).To(Succeed())

			Expect(util.CmpNet(a, b)).NotTo(BeZero())
		})
	})
})
