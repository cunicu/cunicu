package watcher_test

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/watcher"
	"riasc.eu/wice/pkg/wg"
)

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Watcher Suite")
}

var _ = test.SetupLogging()

var _ = BeforeSuite(func() {
	if !util.HasAdminPrivileges() {
		Skip("Insufficient privileges")
	}
})

var _ = Describe("watcher", func() {
	var err error
	var c *wgctrl.Client
	var w *watcher.Watcher
	var devName string
	var user bool
	var h *core.EventsHandler

	test := func() {
		BeforeEach(OncePerOrdered, func() {
			// Generate unique name per test
			devName = fmt.Sprintf("wg-test-%d", rand.Intn(1000))

			c, err = wgctrl.New()
			Expect(err).To(Succeed())

			w, err = watcher.New(c, time.Hour, func(dev string) bool { return dev == devName })
			Expect(err).To(Succeed())

			h = core.NewEventsHandler(16)

			w.OnAll(h)

			go w.Run()
		})

		AfterEach(OncePerOrdered, func() {
			err := w.Close()
			Expect(err).To(Succeed())
		})

		Describe("can watch changes", Ordered, func() {
			var i *core.Interface
			var p *core.Peer
			var d device.Device

			It("adding interface", func() {
				d, err = device.NewDevice(devName, user)
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var ie core.InterfaceAddedEvent
				Expect(h.Events).To(test.ReceiveEvent(&ie))

				i = ie.Interface

				Expect(ie.Interface).NotTo(BeNil())
				Expect(ie.Interface.Name()).To(Equal(devName))
			})

			It("modifying the interface", func() {
				oldListenPort := i.ListenPort
				newListenPort := oldListenPort + 1

				err = i.Configure(&wg.Config{
					Config: wgtypes.Config{
						ListenPort: &newListenPort,
					},
				})
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var ie core.InterfaceModifiedEvent
				Expect(h.Events).To(test.ReceiveEvent(&ie))

				Expect(ie.Interface).NotTo(BeNil())
				Expect(ie.Interface.Name()).To(Equal(devName))
				Expect(ie.Modified & core.InterfaceModifiedListenPort).NotTo(BeZero())
				Expect(ie.Interface.ListenPort).To(Equal(newListenPort))
				Expect(ie.Old.ListenPort).To(Equal(oldListenPort))
			})

			It("adding a peer", func() {
				sk, err := crypto.GenerateKey()
				Expect(err).To(Succeed())

				err = i.AddPeer(&wgtypes.PeerConfig{
					PublicKey: wgtypes.Key(sk.PublicKey()),
				})
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var ie core.InterfaceModifiedEvent
				Expect(h.Events).To(test.ReceiveEvent(&ie))

				Expect(ie.Modified & core.InterfaceModifiedPeers).NotTo(BeZero())

				var pe core.PeerAddedEvent
				Expect(h.Events).To(test.ReceiveEvent(&pe))

				// Remember peer for further test subjects
				p = pe.Peer

				Expect(pe.Peer.PublicKey()).To(Equal(sk.PublicKey()))
			})

			It("modifying a peer", func() {
				ep, err := net.ResolveUDPAddr("udp", "1.1.1.1:1234")
				Expect(err).To(Succeed())

				err = p.UpdateEndpoint(ep)
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var pe core.PeerModifiedEvent
				Expect(h.Events).To(test.ReceiveEvent(&pe))

				Expect(pe.Peer.PublicKey()).To(Equal(p.PublicKey()))
				Expect(pe.Modified & core.PeerModifiedEndpoint).NotTo(BeZero())
				Expect(pe.Peer.Endpoint.IP.Equal(ep.IP)).To(BeTrue())
				Expect(pe.Peer.Endpoint.Port).To(Equal(ep.Port))
			})

			It("removing a peer", func() {
				err := i.RemovePeer(p.PublicKey())
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var ie core.InterfaceModifiedEvent
				Expect(h.Events).To(test.ReceiveEvent(&ie))

				Expect(ie.Modified & core.InterfaceModifiedPeers).NotTo(BeZero())

				var pe core.PeerRemovedEvent
				Expect(h.Events).To(test.ReceiveEvent(&pe))

				Expect(pe.Peer.PublicKey()).To(Equal(p.PublicKey()))
			})

			It("removing an interface", func() {
				err = d.Close()
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var ie core.InterfaceRemovedEvent
				Expect(h.Events).To(test.ReceiveEvent(&ie))

				Expect(ie.Interface).NotTo(BeNil())
				Expect(ie.Interface.Name()).To(Equal(devName))
			})
		})
	}

	Describe("kernel", func() {
		BeforeEach(OncePerOrdered, func() {
			user = false
		})

		test()
	})

	// Describe("user", func() {
	// 	BeforeEach(OncePerOrdered, func() {
	// 		user = true
	// 	})

	// 	test()
	// })
})
