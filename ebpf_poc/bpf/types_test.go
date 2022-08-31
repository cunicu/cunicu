package bpf_test

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"riasc.eu/wice/tc_test/bpf"

	"github.com/cilium/ebpf"
	"github.com/florianl/go-tc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	o "github.com/stv0g/gont/pkg/options"
	"github.com/vishvananda/netlink"
)

func logTShark(hs ...*g.Host) {
	for _, h := range hs {
		stdout, _, _, err := h.Start("tshark", "-o", "udp.check_checksum:TRUE", "-o", "ip.check_checksum:TRUE", "-i", "eth0", "-V", "-x")
		Expect(err).To(Succeed())

		fn := fmt.Sprintf("tshark_%s.log", h.Name())
		f, _ := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		go io.Copy(f, stdout)
	}

	time.Sleep(2 * time.Second)
}

var _ = Describe("bpf", Ordered, func() {
	var err error
	var n *g.Network
	var h1, h2 *g.Host

	var objs *bpf.Objects

	BeforeAll(func() {
		n, err = g.NewNetwork("", o.Persistent(true))
		Expect(err).To(Succeed())

		h1, err = n.AddHost("h1")
		Expect(err).To(Succeed())

		h2, err = n.AddHost("h2")
		Expect(err).To(Succeed())

		err = n.AddLink(
			o.Interface("eth0", h1,
				o.AddressIP("10.0.0.1/24"),
			),
			o.Interface("eth0", h2,
				o.AddressIP("10.0.0.2/24"),
			),
		)

		Expect(err).To(Succeed())
	})

	It("can ping between the hosts", func() {
		stats, err := h1.Ping(h2)
		Expect(err).To(Succeed())

		Expect(stats.MaxRtt).To(BeNumerically("<", 10*time.Millisecond))
	})

	Describe("ebpf", Ordered, func() {
		BeforeAll(func() {
			objs, err = bpf.Load()
			Expect(err).To(Succeed())

			objs.Maps.SettingsMap.EnableDebug()
		})

		AfterAll(func() {
			Expect(objs.Close()).To(Succeed())
		})

		It("has loaded properly", func() {
			Expect(objs.Programs.EgressFilter.Type()).To(Equal(ebpf.SchedCLS))
		})

		Context("maps", Ordered, func() {
			var addr = &net.UDPAddr{
				IP:   net.ParseIP("1.2.3.4"),
				Port: 1234,
			}

			var me = &bpf.MapStateEntry{
				ChannelId: 1,
				Lport:     2222,
			}

			It("can put an entry in the map", func() {
				Expect(objs).NotTo(BeNil())

				err := objs.Maps.EgressMap.AddEntry(addr, me)
				Expect(err).To(Succeed())
			})

			It("should fail to add the same entry again", func() {
				err := objs.Maps.EgressMap.AddEntry(addr, &bpf.MapStateEntry{
					ChannelId: 2,
					Lport:     3333,
				})
				Expect(err).To(HaveOccurred())
			})

			It("can retrieve the entry again", func() {
				me2, err := objs.Maps.EgressMap.GetEntry(addr)
				Expect(err).To(Succeed())

				Expect(me).To(Equal(me2))
			})

			It("can delete an entry from the map", func() {
				err := objs.Maps.EgressMap.DeleteEntry(addr)
				Expect(err).To(Succeed())
			})

			It("is not in the map after the delete", func() {
				_, err := objs.Maps.EgressMap.GetEntry(addr)
				Expect(err).To(MatchError(ebpf.ErrKeyNotExist))
			})
		})

		Context("egress filtering", Ordered, func() {
			var tcnl *tc.Tc
			var link netlink.Link

			var addr = &net.UDPAddr{
				Port: 1234,
			}

			var me = &bpf.MapStateEntry{
				ChannelId: 0,
				Lport:     2222,
			}

			var bufSend []byte

			BeforeAll(func() {
				addr.IP = h2.Interfaces[1].Addresses[0].IP
			})

			BeforeAll(func() {
				// Prepare some test data
				bufSend = make([]byte, 128)
				for i := 0; i < cap(bufSend); i++ {
					bufSend[i] = byte(i)
				}
			})

			// Attach filters
			BeforeAll(func() {
				link, err = h1.NetlinkHandle().LinkByName("eth0")
				Expect(err).To(Succeed(), "could not get interface ID: %v\n", err)

				tcnl, err = tc.Open(&tc.Config{
					NetNS: int(h1.NsHandle),
				})
				Expect(err).To(Succeed(), "Failed to open rtnetlink socket: %v\n", err)

				err := bpf.AttachTCFilters(tcnl, link.Attrs().Index, objs)
				Expect(err).To(Succeed())
			})

			AfterAll(func() {
				Expect(tcnl.Close()).To(Succeed())
			})

			// Configure maps
			BeforeAll(func() {
				Expect(objs.Maps.EgressMap.AddEntry(addr, me)).To(Succeed())
			})

			// AfterAll(func() {
			// 	Expect(objs.Maps.EgressMap.DeleteEntry(addr)).To(Succeed())
			// })

			It("still ping", func() {
				stats, err := h1.Ping(h2)
				Expect(err).To(Succeed())

				Expect(stats.MaxRtt).To(BeNumerically("<", 10*time.Millisecond))
			})

			It("performs a port redirect", func() {
				done := make(chan any)
				listening := make(chan any)

				// logTShark(h1, h2)
				fmt.Scanln()
				time.Sleep(6 * time.Second)

				go h2.RunFunc(func() error {
					defer GinkgoRecover()

					conn, err := net.ListenUDP("udp4", &net.UDPAddr{
						Port: 2222,
					})
					Expect(err).To(Succeed())

					close(listening)

					bufRecv := make([]byte, 128)
					_, _, err = conn.ReadFrom(bufRecv)
					Expect(err).To(Succeed())

					Expect(bufRecv).To(Equal(bufSend))

					close(done)
					return nil
				})

				time.Sleep(time.Second)

				h1.RunFunc(func() error {
					conn, err := net.DialUDP("udp4", &net.UDPAddr{Port: 52722}, &net.UDPAddr{
						Port: 1234,
						IP:   net.ParseIP("10.0.0.2"),
					})
					Expect(err).To(Succeed())

					_, err = conn.Write(bufSend)
					Expect(err).To(Succeed())

					return nil
				})

				Eventually(done).WithTimeout(10 * time.Second).Should(BeClosed())
			})
		})
	})

	Context("egress filtering with TURN channel data indication prefix", Ordered, func() {
		var tcnl *tc.Tc
		var link netlink.Link

		var addr = &net.UDPAddr{
			Port: 1234,
		}

		var me = &bpf.MapStateEntry{
			ChannelId: 0xABAB,
			Lport:     2222,
		}

		var bufSend []byte
		var bufExpect []byte

		BeforeAll(func() {
			addr.IP = h2.Interfaces[1].Addresses[0].IP
		})

		BeforeAll(func() {
			// Prepare some test data
			bufSend = make([]byte, 128)
			for i := 0; i < cap(bufSend); i++ {
				bufSend[i] = byte(i)
			}

			bufExpect = make([]byte, 128+4)
			binary.BigEndian.PutUint16(bufExpect[0:], me.ChannelId)
			binary.BigEndian.PutUint16(bufExpect[2:], 0xCCDD)
			for i := 4; i < cap(bufSend); i++ {
				bufExpect[i] = byte(i - 4)
			}
		})

		// Attach filters
		BeforeAll(func() {
			link, err = h1.NetlinkHandle().LinkByName("eth0")
			Expect(err).To(Succeed(), "could not get interface ID: %v\n", err)

			tcnl, err = tc.Open(&tc.Config{
				NetNS: int(h1.NsHandle),
			})
			Expect(err).To(Succeed(), "Failed to open rtnetlink socket: %v\n", err)

			err := bpf.AttachTCFilters(tcnl, link.Attrs().Index, objs)
			Expect(err).To(Succeed())
		})

		AfterAll(func() {
			Expect(tcnl.Close()).To(Succeed())
		})

		// Configure maps
		BeforeAll(func() {
			Expect(objs.Maps.EgressMap.AddEntry(addr, me)).To(Succeed())
		})

		// AfterAll(func() {
		// 	Expect(objs.Maps.EgressMap.DeleteEntry(addr)).To(Succeed())
		// })

		It("still ping", func() {
			stats, err := h1.Ping(h2)
			Expect(err).To(Succeed())

			Expect(stats.MaxRtt).To(BeNumerically("<", 10*time.Millisecond))
		})

		It("performs a port redirect", func() {
			done := make(chan any)
			listening := make(chan any)

			// logTShark(h1, h2)
			fmt.Scanln()
			time.Sleep(6 * time.Second)

			go h2.RunFunc(func() error {
				defer GinkgoRecover()

				conn, err := net.ListenUDP("udp4", &net.UDPAddr{
					Port: 2222,
				})
				Expect(err).To(Succeed())

				close(listening)

				bufRecv := make([]byte, 128)
				_, _, err = conn.ReadFrom(bufRecv)
				Expect(err).To(Succeed())

				Expect(bufRecv).To(Equal(bufExpect))

				close(done)
				return nil
			})

			time.Sleep(time.Second)

			h1.RunFunc(func() error {
				conn, err := net.DialUDP("udp4", &net.UDPAddr{Port: 52722}, &net.UDPAddr{
					Port: 1234,
					IP:   net.ParseIP("10.0.0.2"),
				})
				Expect(err).To(Succeed())

				_, err = conn.Write(bufSend)
				Expect(err).To(Succeed())

				return nil
			})

			Eventually(done).WithTimeout(10 * time.Second).Should(BeClosed())
		})
	})

	AfterAll(func() {
		Expect(n.Close()).To(Succeed())
	})
})
