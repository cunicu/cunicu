//go:build linux

package watcher_test

import (
	"fmt"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/stv0g/gont/pkg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/watcher"
	"riasc.eu/wice/pkg/wg"
	"riasc.eu/wice/test"
)

var _ = Describe("watcher", func() {
	var err error
	var c *wgctrl.Client
	var w *watcher.Watcher
	var user bool
	var h *core.EventsHandler
	var ns *g.Namespace

	JustBeforeEach(OncePerOrdered, func() {
		name := fmt.Sprintf("wg-test-%d", rand.Intn(1000))
		ns, err = g.NewNamespace(name)
		Expect(err).To(Succeed())

		func() {
			exit, err := ns.Enter()
			Expect(err).To(Succeed())

			defer exit()

			c, err = wgctrl.New()
			Expect(err).To(Succeed())
		}()

		w, err = watcher.New(c, time.Hour, func(dev string) bool { return dev == "wg0" })
		Expect(err).To(Succeed())

		h = core.NewEventsHandler(16)

		w.OnAll(h)

		go w.Run()
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

	Test := func() {
		Describe("can watch changes", Ordered, func() {
			var i *core.Interface
			var p *core.Peer
			var d device.Device

			It("adding interface", func() {
				d, err = device.NewDevice("wg0", user)
				Expect(err).To(Succeed())

				err = w.Sync()
				Expect(err).To(Succeed())

				var ie core.InterfaceAddedEvent
				Expect(h.Events).To(test.ReceiveEvent(&ie))

				i = ie.Interface

				Expect(ie.Interface).NotTo(BeNil())
				Expect(ie.Interface.Name()).To(Equal("wg0"))
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
				Expect(ie.Interface.Name()).To(Equal("wg0"))
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
				Expect(ie.Interface.Name()).To(Equal("wg0"))
			})
		})
	}

	Describe("kernel", func() {
		BeforeEach(OncePerOrdered, func() {
			user = false
		})

		Test()
	})

	Describe("user", func() {
		BeforeEach(OncePerOrdered, func() {
			user = true
		})

		Test()
	})
})
