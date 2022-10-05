//go:build linux

package watcher_test

import (
	"fmt"
	"math/rand"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/test"
	g "github.com/stv0g/gont/pkg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var _ = Describe("watcher", func() {
	var err error
	var c *wgctrl.Client
	var w *watcher.Watcher
	var devUser bool
	var h *core.EventsHandler
	var ns *g.Namespace
	var devName string

	JustBeforeEach(OncePerOrdered, func() {
		By("Creating net namespace")

		nsName := fmt.Sprintf("wg-test-ns-%d", rand.Intn(1000))
		ns, err = g.NewNamespace(nsName)
		Expect(err).To(Succeed())

		func() {
			exit, err := ns.Enter()
			Expect(err).To(Succeed())

			defer exit()

			c, err = wgctrl.New()
			Expect(err).To(Succeed())
		}()

		By("Creating watcher")

		w, err = watcher.New(c, 0, func(dev string) bool { return dev == devName })
		Expect(err).To(Succeed())

		h = core.NewEventsHandler(16)

		w.OnAll(h)

		go w.Watch()
	})

	JustAfterEach(OncePerOrdered, func() {
		err := w.Close()
		Expect(err).To(Succeed())

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

	TestSync := func() {
		var i *core.Interface
		var p *core.Peer
		var d device.Device

		It("adding interface", func() {
			d, err = device.NewDevice(devName, devUser)
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

			err = c.ConfigureDevice(devName, wgtypes.Config{
				ListenPort: &newListenPort,
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

			err = c.ConfigureDevice(i.Name(), wgtypes.Config{
				Peers: []wgtypes.PeerConfig{
					{
						PublicKey: wgtypes.Key(sk.PublicKey()),
					},
				},
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
			psk, err := crypto.GeneratePrivateKey()
			Expect(err).To(Succeed())

			err = p.SetPresharedKey(&psk)
			Expect(err).To(Succeed())

			err = w.Sync()
			Expect(err).To(Succeed())

			var pe core.PeerModifiedEvent
			Expect(h.Events).To(test.ReceiveEvent(&pe))

			Expect(pe.Peer.PublicKey()).To(Equal(p.PublicKey()))
			Expect(pe.Modified & core.PeerModifiedPresharedKey).NotTo(BeZero())
			Expect(pe.Peer.PresharedKey()).To(Equal(psk))
		})

		It("removing a peer", func() {
			err = c.ConfigureDevice(i.Name(), wgtypes.Config{
				Peers: []wgtypes.PeerConfig{
					{
						PublicKey: wgtypes.Key(p.PublicKey()),
						Remove:    true,
					},
				},
			})
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
	}

	Context("sync", func() {
		Context("kernel", Ordered, func() {
			BeforeAll(func() {
				devName = "wg-kernel"
				devUser = false
			})

			TestSync()
		})

		Context("user", Ordered, func() {
			BeforeAll(func() {
				devName = fmt.Sprintf("wg-user-%d", rand.Intn(1000))
				devUser = true
			})

			TestSync()
		})
	})

	Context("watch", Pending, func() {
		Context("periodic", func() {

		})

		Context("user", func() {

		})

		Context("kernel", func() {

		})
	})
})
