// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package link_test

import (
	"fmt"
	"math/rand"
	"net"
	"syscall"

	g "github.com/stv0g/gont/v2/pkg"
	nl "github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
	wgdevice "golang.zx2c4.com/wireguard/device"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/device"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("device", func() {
	var err error
	var d device.Device
	var user bool
	var ns *g.Namespace
	var nlh *nl.Handle
	var l nl.Link

	getAddrs := func() []net.IPNet {
		addrs, err := nlh.AddrList(l, unix.AF_INET)
		Expect(err).To(Succeed())

		ips := []net.IPNet{}
		for _, addr := range addrs {
			addr.IP = addr.IP.To16()
			ips = append(ips, *addr.IPNet)
		}

		return ips
	}

	JustBeforeEach(OncePerOrdered, func() {
		name := fmt.Sprintf("wg-test-%d", rand.Intn(1000)) //nolint:gosec
		ns, err = g.NewNamespace(name)
		Expect(err).To(Succeed())

		func() {
			exit, err := ns.Enter()
			defer exit()

			Expect(err).To(Succeed())

			d, err = device.NewDevice("wg0", user)
			Expect(err).To(Succeed())

			nlh, err = nl.NewHandleAt(ns.NsHandle)
			Expect(err).To(Succeed())

			l, err = nlh.LinkByName("wg0")
			Expect(err).To(Succeed())
		}()
	})

	JustAfterEach(OncePerOrdered, func() {
		func() {
			exit, err := ns.Enter()
			Expect(err).To(Succeed())

			defer exit()

			err = d.Close()
			Expect(err).To(Succeed())
		}()

		err = ns.Close()
		Expect(err).To(Succeed())
	})

	// Overload It() to enter/exit namespace
	It := func(name string, function func()) {
		It(name, Offset(1), func() {
			exit, err := ns.Enter()
			Expect(err).To(Succeed())

			function()
			exit()
		})
	}

	Test := func() {
		Context("", Ordered, func() {
			It("have a matching name", func() {
				Expect(d.Name()).To(Equal("wg0"))
			})

			Describe("can manage addresses", Ordered, func() {
				var addr net.IPNet

				BeforeAll(func() {
					addr = net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0xff),
					}
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
				var ip net.IP
				var route net.IPNet

				BeforeAll(func() {
					ip = net.IPv4(10, 0, 0, 1)

					route = net.IPNet{
						IP:   net.IPv4(10, 0, 0, 1),
						Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0xff),
					}
				})

				It("can add a new route", func() {
					// The device must be up before adding device routes
					err = d.SetUp()
					Expect(err).To(Succeed())

					err = d.AddRoute(route, nil, config.DefaultRouteTable)
					Expect(err).To(Succeed())

					routes, err := nl.RouteGet(ip)
					Expect(err).To(Succeed())
					Expect(routes).To(HaveLen(1))
					Expect(routes[0].LinkIndex).To(Equal(d.Index()))
				})

				It("can delete the route again", func() {
					err = d.DeleteRoute(route, config.DefaultRouteTable)
					Expect(err).To(Succeed())

					_, err := nl.RouteGet(ip)
					Expect(err).To(MatchError(syscall.ENETUNREACH))
				})
			})

			Describe("can change link state", Ordered, func() {
				isUP := func() bool {
					l, err = nlh.LinkByName("wg0")
					if err != nil {
						return false
					}

					return l.Attrs().Flags&net.FlagUp != 0
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

			Describe("can change link attributes", Ordered, func() {
				It("has the correct interface index", func() {
					Expect(d.Index()).To(Equal(l.Attrs().Index))
				})

				It("has the default MTU initially", func() {
					Expect(d.MTU()).To(Equal(wgdevice.DefaultMTU))
				})

				It("has the correct MTU", func() {
					Expect(d.MTU()).To(Equal(l.Attrs().MTU))
				})

				It("updates the MTU", func() {
					newMTU := 1300

					err = d.SetMTU(newMTU)
					Expect(err).To(Succeed())

					Expect(d.MTU()).To(Equal(newMTU))

					// Get updated link attributes
					l, err = nlh.LinkByName("wg0")
					Expect(err).To(Succeed())

					Expect(l.Attrs().MTU).To(Equal(newMTU))
				})
			})
		})
	}

	When("kernel", Ordered, func() {
		BeforeEach(OncePerOrdered, func() {
			user = false
		})

		Test()
	})

	When("user", func() {
		BeforeEach(OncePerOrdered, func() {
			user = true
		})

		Test()
	})
})
