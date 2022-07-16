package core_test

import (
	"fmt"
	"math/rand"
	"net"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/core"
)

var _ = Describe("interface", func() {
	var err error
	var i *core.Interface
	var c *wgctrl.Client
	var intfName string
	var user bool

	test := func() {
		It("have a matching name", func() {
			Expect(i.Name()).To(Equal(intfName))
		})

		// Describe("config sync", func() {
		// 	var cfgPath string
		// 	var privKey crypto.Key
		// 	var listenPort int

		// 	BeforeEach(func() {
		// 		cfgPath = GinkgoT().TempDir()
		// 		listenPort = config.WireguardDefaultPort + rand.Intn(1000)
		// 		privKey, err = crypto.GeneratePrivateKey()
		// 		Expect(err).To(Succeed())

		// 		cfg.Settings.Wireguard.Config.Sync = true
		// 		cfg.Settings.Wireguard.Config.Path = cfgPath

		// 		wgCfg := wg.Config{
		// 			Config: wgtypes.Config{
		// 				PrivateKey: (*wgtypes.Key)(&privKey),
		// 				ListenPort: &listenPort,
		// 			},
		// 		}

		// 		f, err := os.OpenFile(path.Join(cfgPath, fmt.Sprintf("%s.conf", intfName)), os.O_CREATE|os.O_WRONLY, 0755)
		// 		Expect(err).To(Succeed())

		// 		err = wgCfg.Dump(f)
		// 		Expect(err).To(Succeed())
		// 	})

		// 	It("key matches", func() {
		// 		Expect(i.PrivateKey).To(Equal(privKey))
		// 	})

		// 	It("listen port matches", func() {
		// 		Expect(i.ListenPort).To(Equal(listenPort))
		// 	})
		// })

		Describe("check attributes", func() {
			var j *net.Interface

			JustBeforeEach(func() {
				j, err = net.InterfaceByName(intfName)
				Expect(err).To(Succeed())
			})

			It("is down", func() {
				Expect(j.Flags & net.FlagUp).To(BeZero())
			})

			It("has the correct interface index", func() {
				Expect(i.KernelDevice.Index()).To(Equal(j.Index))
			})

			It("has the default MTU initially", func() {
				Expect(i.KernelDevice.MTU()).To(Equal(device.DefaultMTU))
			})

			It("can set the MTU", func() {
				err = i.KernelDevice.SetMTU(1300)
				Expect(err).To(Succeed())

				Expect(i.KernelDevice.MTU()).To(Equal(1300))
			})
		})
	}

	JustBeforeEach(OncePerOrdered, func() {
		// Get wg client
		c, err = wgctrl.New()
		Expect(err).To(Succeed())

		// Generate unique name per test
		intfName = fmt.Sprintf("wg-test-%d", rand.Intn(1000))

		i, err = core.CreateInterface(intfName, user, c)
		Expect(err).To(Succeed())
	})

	AfterEach(OncePerOrdered, func() {
		err = i.Close()
		Expect(err).To(Succeed())
	})

	When("using a kernel interface", func() {
		BeforeEach(OncePerOrdered, func() {
			user = false
		})

		test()
	})

	When("using a userspace interface", func() {
		BeforeEach(OncePerOrdered, func() {
			user = true
		})

		test()
	})
})
