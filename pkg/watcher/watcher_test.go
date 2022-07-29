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
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/test"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/watcher"
)

func TestSuite(t *testing.T) {
	rand.Seed(GinkgoRandomSeed())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Watcher Suite")
}

var _ = test.SetupLogging()

var _ = BeforeSuite(func() {
	if !util.HasCapabilities(cap.NET_ADMIN) {
		Skip("Insufficient privileges")
	}
})

var _ = Describe("watcher", func() {
	var err error
	var c *wgctrl.Client
	var w *watcher.Watcher
	var devName string
	var newDevice func(string) (device.KernelDevice, error)
	var h *core.EventsHandler

	BeforeEach(OncePerOrdered, func() {
		// Generate unique name per test
		devName = fmt.Sprintf("wg-test-%d", rand.Intn(1000))

		c, err = wgctrl.New()
		Expect(err).To(Succeed())

		w, err = watcher.New(c, time.Second, nil)
		Expect(err).To(Succeed())

		h = core.NewMockHandler()

		w.OnAll(h)

		go w.Run()
	})

	AfterEach(OncePerOrdered, func() {
		err := w.Close()
		Expect(err).To(Succeed())
	})

	test := func() {
		Describe("can watch changes", Ordered, func() {
			var i *core.Interface
			var p *core.Peer
			var d device.KernelDevice

			It("adding interface", func() {
				d, err = newDevice(devName)
				Expect(err).To(Succeed())

				Eventually(func(g Gomega) {
					var ie core.InterfaceAddedEvent
					g.Eventually(h.Events).Should(Receive(&ie))

					// Remember interface for further test subjects
					i = ie.Interface

					g.Expect(ie.Interface).NotTo(BeNil())
					g.Expect(ie.Interface.Name()).To(Equal(devName))
				}).Should(Succeed())
			})

			It("modifying the interface", func() {
				oldListenPort := i.ListenPort
				newListenPort := oldListenPort + 1

				err = i.Configure(wgtypes.Config{
					ListenPort: &newListenPort,
				})
				Expect(err).To(Succeed())

				Eventually(func(g Gomega) {
					var ie core.InterfaceModifiedEvent
					g.Eventually(h.Events).Should(Receive(&ie))

					g.Expect(ie.Modified & core.InterfaceModifiedListenPort).NotTo(BeZero())
					g.Expect(ie.Interface.ListenPort).To(Equal(newListenPort))
					g.Expect(ie.Old.ListenPort).To(Equal(oldListenPort))
				}).Should(Succeed())
			})

			It("adding a peer", func() {
				sk, err := crypto.GenerateKey()
				Expect(err).To(Succeed())

				err = i.AddPeer(sk.PublicKey())
				Expect(err).To(Succeed())

				Eventually(func(g Gomega) {
					var ie core.InterfaceModifiedEvent
					g.Eventually(h.Events).Should(Receive(&ie))

					g.Expect(ie.Modified & core.InterfaceModifiedPeers).NotTo(BeZero())
				}).Should(Succeed())

				Eventually(func(g Gomega) {
					var ie core.PeerAddedEvent
					g.Eventually(h.Events).Should(Receive(&ie))

					// Remember peer for further test subjects
					p = ie.Peer

					g.Expect(ie.Peer.PublicKey()).To(Equal(sk.PublicKey()))
				}).Should(Succeed())
			})

			It("modifying a peer", func() {
				ep, err := net.ResolveUDPAddr("udp", "1.1.1.1:1234")
				Expect(err).To(Succeed())

				err = p.UpdateEndpoint(ep)
				Expect(err).To(Succeed())

				Eventually(func(g Gomega) {
					var ie core.PeerModifiedEvent
					g.Eventually(h.Events).Should(Receive(&ie))

					g.Expect(ie.Modified & core.PeerModifiedEndpoint).NotTo(BeZero())
					g.Expect(ie.Peer.Endpoint.IP.Equal(ep.IP)).To(BeTrue())
					g.Expect(ie.Peer.Endpoint.Port).To(Equal(ep.Port))
				}).Should(Succeed())
			})

			It("removing a peer", func() {

			})

			It("removing an interface", func() {
				err = d.Close()
				Expect(err).To(Succeed())

				// Eventually(func() int { return h.interfaceCount }).Should(BeNumerically("==", 0))
			})
		})
	}

	Describe("kernel", func() {
		BeforeEach(func() {
			newDevice = device.NewKernelDevice
		})

		test()
	})

	Describe("user", func() {
		BeforeEach(func() {
			newDevice = device.NewUserDevice
		})

		test()
	})
})
