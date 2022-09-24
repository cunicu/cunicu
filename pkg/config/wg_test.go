package config_test

import (
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/config"
)

var _ = Context("wg provider", func() {
	var dir string

	createConfigFile := func(name, contents string) {
		err := os.WriteFile(filepath.Join(dir, name+".conf"), []byte(contents), 0644)
		Expect(err).To(Succeed())
	}

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
		os.Setenv("WG_CONFIG_PATH", dir)
	})

	AfterEach(func() {
		os.Unsetenv("WG_CONFIG_PATH")
	})

	It("can parse a directory of WireGuard config files", func() {
		createConfigFile("wg0", `
[Interface]
Address = 10.200.100.8/24
DNS = 10.200.100.1
Table = 123
MTU = 1380
FwMark = 0x1000
PrivateKey = oK56DE9Ue9zK76rAc8pBl6opph+1v36lm7cXXsQKrQM=
		
[Peer]
PublicKey = GtL7fZc/bLnqZldpVofMCD6hDjrK28SsdLxevJ+qtKU=
PresharedKey = /UwcSPg38hW/D9Y3tcS1FOV0K1wuURMbS0sesJEP5ak=
AllowedIPs = 0.0.0.0/0
Endpoint = localhost:51820
PersistentKeepalive = 25
`)

		createConfigFile("wg1", `
[Interface]
PrivateKey = mBVQEpzmRVRRkba82CorTcbE2Zab4KhAtlNhDm4DYXo=
`)

		cfg, err := config.ParseArgs()
		Expect(err).To(Succeed())

		Expect(cfg.InterfaceOrder).To(ContainElements("wg0", "wg1"))
		Expect(cfg.Interfaces).To(HaveLen(2))

		icfg1 := cfg.InterfaceSettings("wg0")
		Expect(icfg1).NotTo(BeNil())

		Expect(icfg1.WireGuard.PrivateKey.String()).To(Equal("oK56DE9Ue9zK76rAc8pBl6opph+1v36lm7cXXsQKrQM="))
		Expect(icfg1.WireGuard.FirewallMark).To(Equal(0x1000))
		Expect(icfg1.RouteSync.Table).To(Equal(123))
		Expect(icfg1.AutoConfig.Addresses).To(HaveLen(1))
		Expect(icfg1.AutoConfig.Addresses[0].String()).To(Equal("10.200.100.8/24"))
		Expect(icfg1.AutoConfig.MTU).To(Equal(1380))
		Expect(icfg1.AutoConfig.DNS).To(HaveLen(1))
		Expect(icfg1.AutoConfig.DNS[0].String()).To(Equal("10.200.100.1"))

		Expect(icfg1.WireGuard.Peers).To(HaveLen(1))
		Expect(icfg1.WireGuard.Peers[0].PublicKey.String()).To(Equal("GtL7fZc/bLnqZldpVofMCD6hDjrK28SsdLxevJ+qtKU="))
		Expect(icfg1.WireGuard.Peers[0].Endpoint).To(Equal("localhost:51820"))
		Expect(icfg1.WireGuard.Peers[0].PersistentKeepaliveInterval).To(Equal(25 * time.Second))

		icfg2 := cfg.InterfaceSettings("wg1")
		Expect(icfg2).NotTo(BeNil())

		Expect(icfg2.WireGuard.PrivateKey.String()).To(Equal("mBVQEpzmRVRRkba82CorTcbE2Zab4KhAtlNhDm4DYXo="))
		Expect(icfg2.WireGuard.Peers).To(HaveLen(0))
	})
})
