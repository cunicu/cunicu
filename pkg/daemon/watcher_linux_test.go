// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package daemon_test

import (
	"fmt"
	"math/rand"

	g "github.com/stv0g/gont/v2/pkg"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("watcher", func() {
	var err error
	var c *wgctrl.Client
	var w *daemon.Watcher
	var devUser bool
	var h *daemon.EventsHandler
	var ns *g.Namespace
	var devName string

	JustBeforeEach(OncePerOrdered, func() {
		By("Creating net namespace")

		nsName := fmt.Sprintf("wg-test-ns-%d", rand.Intn(1000)) //nolint:gosec
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

		w, err = daemon.NewWatcher(c, 0, func(dev string) bool { return dev == devName })
		Expect(err).To(Succeed())

		h = daemon.NewEventsHandler(16)

		w.AddAllHandler(h)

		go func() {
			exit, _ := ns.Enter()
			defer exit()

			w.Watch()
		}()
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
		var i *daemon.Interface
		var p *daemon.Peer
		var d device.Device

		It("adding interface", func() {
			d, err = device.NewDevice(devName, devUser)
			Expect(err).To(Succeed())

			err = w.Sync()
			Expect(err).To(Succeed())

			var ie daemon.InterfaceAddedEvent
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

			var ie daemon.InterfaceModifiedEvent
			Expect(h.Events).To(test.ReceiveEvent(&ie))

			Expect(ie.Interface).NotTo(BeNil())
			Expect(ie.Interface.Name()).To(Equal(devName))
			Expect(ie.Modified & daemon.InterfaceModifiedListenPort).NotTo(BeZero())
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

			var ie daemon.InterfaceModifiedEvent
			Expect(h.Events).To(test.ReceiveEvent(&ie))

			Expect(ie.Modified & daemon.InterfaceModifiedPeers).NotTo(BeZero())

			var pe daemon.PeerAddedEvent
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

			var pe daemon.PeerModifiedEvent
			Expect(h.Events).To(test.ReceiveEvent(&pe))

			Expect(pe.Peer.PublicKey()).To(Equal(p.PublicKey()))
			Expect(pe.Modified & daemon.PeerModifiedPresharedKey).NotTo(BeZero())
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

			var ie daemon.InterfaceModifiedEvent
			Expect(h.Events).To(test.ReceiveEvent(&ie))

			Expect(ie.Modified & daemon.InterfaceModifiedPeers).NotTo(BeZero())

			var pe daemon.PeerRemovedEvent
			Expect(h.Events).To(test.ReceiveEvent(&pe))

			Expect(pe.Peer.PublicKey()).To(Equal(p.PublicKey()))
		})

		It("removing an interface", func() {
			err = d.Close()
			Expect(err).To(Succeed())

			err = w.Sync()
			Expect(err).To(Succeed())

			var ie daemon.InterfaceRemovedEvent
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
				devName = fmt.Sprintf("wg-user-%d", rand.Intn(1000)) //nolint:gosec
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
