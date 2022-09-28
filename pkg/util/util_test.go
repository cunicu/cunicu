package util_test

import (
	"math/rand"
	"net"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/test"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utilities Suite")
}

var _ = test.SetupLogging()

var _ = Context("endpoint comparisons", func() {
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

var _ = Context("network comparisons", func() {
	It("compare equal networks", func() {
		_, a, err := net.ParseCIDR("1.1.1.1/0")
		Expect(err).To(Succeed())

		Expect(util.CmpNet(*a, *a)).To(BeZero())
	})

	It("compare unequal networks", func() {
		_, a, err := net.ParseCIDR("1.1.1.1/0")
		Expect(err).To(Succeed())

		_, b, err := net.ParseCIDR("1.1.1.1/1")
		Expect(err).To(Succeed())

		Expect(util.CmpNet(*a, *b)).NotTo(BeZero())
	})
})

var _ = Context("contains net", func() {
	_, net1, _ := net.ParseCIDR("1.1.1.1/24")
	_, net2, _ := net.ParseCIDR("1.1.0.2/16")
	_, net3, _ := net.ParseCIDR("1.1.1.3/25")
	_, net4, _ := net.ParseCIDR("1.2.0.4/16")

	DescribeTable("network intersection",
		func(outer, inner *net.IPNet, expect bool) {
			result := util.ContainsNet(outer, inner)
			Expect(result).To(Equal(expect), "Expected a=%s, b=%s, expected=%t, result=%t", outer, inner, result, expect)
		},
		Entry("net1 contains net2 = true", net1, net2, false),
		Entry("net2 contains net1 = true", net2, net1, true),
		Entry("net1 contains net3 = true", net1, net3, true),
		Entry("net3 contains net1 = true", net3, net1, false),
		Entry("net1 contains net4 = false", net1, net4, false),
		Entry("net4 contains net1 = false", net4, net1, false),
		Entry("net1 contains net1 = true", net1, net1, true),
	)
})

var _ = It("rand", func() {
	a := rand.Float32()

	util.SetupRand()
	b := rand.Float32()

	util.SetupRand()
	c := rand.Float32()

	Expect(a).NotTo(Equal(b))
	Expect(a).NotTo(Equal(c))
	Expect(b).NotTo(Equal(c))
})

var _ = Context("offset ip", func() {
	It("ipv4", func() {
		ip1 := net.ParseIP("1.2.3.4")
		Expect(ip1).NotTo(BeNil())

		ip2 := net.ParseIP("1.2.3.14")
		Expect(ip2).NotTo(BeNil())

		Expect(util.OffsetIP(ip1, 10)).To(Equal(ip2))

	})

	It("ipv6", func() {
		ip1 := net.ParseIP("fc::1:2:3:4")
		Expect(ip1).NotTo(BeNil())

		ip2 := net.ParseIP("fc::1:2:3:e")
		Expect(ip2).NotTo(BeNil())

		Expect(util.OffsetIP(ip1, 10)).To(Equal(ip2))
	})
})
