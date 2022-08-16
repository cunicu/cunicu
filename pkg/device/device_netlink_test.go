//go:build linux

package device_test

import (
	"fmt"
	"math/rand"
	"net"

	wgdevice "golang.zx2c4.com/wireguard/device"
	"riasc.eu/wice/pkg/device"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vishvananda/netlink"
)

func randIPNet() *net.IPNet {
	return &net.IPNet{
		IP:   net.IPv4(10, 0, 0, byte(rand.Intn(255))),
		Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0xff),
	}
}

var _ = Describe("device", func() {
	var err error
	var d device.Device
	var devName string

	getAddrs := func() []*net.IPNet {
		intf, _ := net.InterfaceByName(devName)
		addrs, _ := intf.Addrs()

		ips := []*net.IPNet{}
		for _, addr := range addrs {
			ips = append(ips, addr.(*net.IPNet))
		}

		return ips
	}

	test := func() {
		It("have a matching name", func() {
			Expect(d.Name()).To(Equal(devName))
		})

		Describe("can manage addresses", Ordered, func() {
			var addr *net.IPNet

			BeforeAll(func() {
				addr = randIPNet()
			})

			It("can add a new address", func() {
				err = d.AddAddress(addr)
				Expect(err).To(Succeed())

				addrs := getAddrs()
				Expect(addrs).To(ContainElement(addr))
			})

			It("can delete the address again", func() {
				err = d.DeleteAddress(addr)
				Expect(err).To(Succeed())

				addrs := getAddrs()
				Expect(addrs).To(BeEmpty())
			})
		})

		Describe("can manage routes", Ordered, func() {
			var addr *net.IPNet

			BeforeAll(func() {
				addr = randIPNet()
			})

			It("can add a new route", func() {
				// The device must be up before adding device routes
				err = d.SetUp()
				Expect(err).To(Succeed())

				err = d.AddRoute(addr)
				Expect(err).To(Succeed())

				routes, err := netlink.RouteGet(addr.IP)
				Expect(err).To(Succeed())
				Expect(routes).To(HaveLen(1))
				Expect(routes[0].LinkIndex).To(Equal(d.Index()))
			})

			It("can delete the route again", func() {
				err = d.DeleteRoute(addr)
				Expect(err).To(Succeed())

				routes, err := netlink.RouteGet(addr.IP)
				Expect(err).To(Succeed())
				Expect(routes).To(HaveLen(1))
				Expect(routes[0].LinkIndex).NotTo(Equal(d.Index()))
			})
		})

		Describe("can change link state", Ordered, func() {
			var j *net.Interface

			isUP := func() bool {
				j, err = net.InterfaceByName(devName)
				Expect(err).To(Succeed())

				return j.Flags&net.FlagUp != 0
			}

			It("can be brought up", func() {
				err = d.SetUp()
				Expect(err).To(Succeed())

				Eventually(isUP).Should(BeTrue())
			})

			It("can be brought down again", func() {
				err = d.SetDown()
				Expect(err).To(Succeed())

				Eventually(isUP).Should(BeFalse())
			})
		})

		Describe("check attributes", func() {
			var j *net.Interface

			BeforeEach(func() {
				j, err = net.InterfaceByName(devName)
				Expect(err).To(Succeed())
			})

			It("has the correct interface index", func() {
				Expect(d.Index()).To(Equal(j.Index))
			})

			It("has the default MTU initially", func() {
				Expect(d.MTU()).To(Equal(wgdevice.DefaultMTU))
			})

			It("updates the MTU", func() {
				err = d.SetMTU(1300)
				Expect(err).To(Succeed())

				Expect(d.MTU()).To(Equal(1300))
			})
		})
	}

	BeforeEach(OncePerOrdered, func() {
		// Generate unique name per test
		devName = fmt.Sprintf("wg-test-%d", rand.Intn(1000))
	})

	AfterEach(OncePerOrdered, func() {
		err = d.Close()
		Expect(err).To(Succeed())
	})

	When("using a kernel device", func() {
		BeforeEach(OncePerOrdered, func() {
			d, err = device.NewKernelDevice(devName)
			Expect(err).To(Succeed())
		})

		test()
	})

	When("using a userspace device", func() {
		BeforeEach(OncePerOrdered, func() {
			d, err = device.NewUserDevice(devName)
			Expect(err).To(Succeed())
		})

		test()
	})
})
