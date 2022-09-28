package config_test

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pion/ice/v2"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/test"

	icex "github.com/stv0g/cunicu/pkg/ice"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = test.SetupLogging()

var _ = Context("config", func() {
	mkTempFile := func(contents string) *os.File {
		dir := GinkgoT().TempDir()
		fn := filepath.Join(dir, "cunicu.yaml")

		file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0600)
		Expect(err).To(Succeed())

		defer file.Close()

		_, err = file.Write([]byte(contents))
		Expect(err).To(Succeed())

		return file
	}

	Describe("parse command line arguments", func() {
		It("can parse a boolean argument like wg-userspace", func() {
			cfg, err := config.ParseArgs("--wg-userspace")
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.WireGuard.UserSpace).To(BeTrue())
		})

		It("can parse multiple backends", func() {
			cfg, err := config.ParseArgs("--backend", "k8s", "--backend", "p2p")

			Expect(err).To(Succeed())
			Expect(cfg.Backends).To(HaveLen(2))
			Expect(cfg.Backends[0].Scheme).To(Equal("k8s"))
			Expect(cfg.Backends[1].Scheme).To(Equal("p2p"))
		})

		It("parse an interface list", func() {
			cfg, err := config.ParseArgs("wg0", "wg1")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg2")).To(BeFalse())
		})

		It("parse an interface list with custom filter", func() {
			cfg, err := config.ParseArgs("--interface-filter", "wg2", "--", "wg0", "wg1")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg2")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg2")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg3")).To(BeFalse())
			Expect(cfg.InterfaceSettings("wg3")).To(BeNil())

		})

		It("fails on invalid arguments", func() {
			_, err := config.ParseArgs("--wrong")

			Expect(err).To(MatchError("failed to parse command line flags: unknown flag: --wrong"))
		})

		It("fails on invalid arguments values", func() {
			_, err := config.ParseArgs("--url", ":_")

			Expect(err).To(MatchError(And(
				HavePrefix("failed unmarshal settings"),
				HaveSuffix("missing protocol scheme"),
			)))
		})

		Describe("parse configuration files", func() {
			Context("with a local file", func() {
				var cfgFile *os.File

				BeforeEach(func() {
					cfgFile = mkTempFile("watch_interval: 1337s\n")
				})

				Context("file with explicit path", func() {
					It("can read a single valid local file", func() {
						cfg, err := config.ParseArgs("--config", cfgFile.Name())

						Expect(err).To(Succeed())
						Expect(cfg.WatchInterval).To(Equal(1337 * time.Second))
					})

					Specify("that command line arguments take precedence over settings provided by configuration files", func() {
						cfg, err := config.ParseArgs("--config", cfgFile.Name(), "--watch-interval", "1m")
						Expect(err).To(Succeed())
						Expect(cfg.WatchInterval).To(Equal(time.Minute))
					})
				})

				Context("in search path", func() {
					var oldHomeDir string

					BeforeEach(func() {
						// Move config file into XDG config directory
						homeDir := filepath.Dir(cfgFile.Name())
						configDir := filepath.Join(homeDir, ".config", "cunicu")

						err := os.MkdirAll(configDir, 0755)
						Expect(err).To(Succeed())

						err = os.Rename(
							cfgFile.Name(),
							filepath.Join(configDir, "cunicu.yaml"),
						)
						Expect(err).To(Succeed())

						oldHomeDir = os.Getenv("HOME")

						err = os.Setenv("HOME", homeDir)
						Expect(err).To(Succeed())
					})

					AfterEach(func() {
						err := os.Setenv("HOME", oldHomeDir)
						Expect(err).To(Succeed())
					})

					It("can read a single valid local file", func() {
						cfg, err := config.ParseArgs()

						Expect(err).To(Succeed())
						Expect(cfg.WatchInterval).To(Equal(1337 * time.Second))
					})
				})
			})

			Context("with a remote URL", func() {
				var server *ghttp.Server

				BeforeEach(func() {
					server = ghttp.NewServer()
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/cunicu.yaml"),
							ghttp.RespondWith(http.StatusOK,
								"watch_interval: 1337s\n",
								http.Header{
									"Content-type": []string{"text/yaml"},
								}),
						),
					)
				})

				It("can fetch a valid remote configuration file", func() {
					cfg, err := config.ParseArgs("--config", server.URL()+"/cunicu.yaml")

					Expect(err).To(Succeed())
					Expect(cfg.WatchInterval).To(BeNumerically("==", 1337*time.Second))
				})

				AfterEach(func() {
					//shut down the server between tests
					server.Close()
				})

				It("fails on loading an non-existent remote file", func() {
					_, err := config.ParseArgs("--config", "http://example.com/doesnotexist.yaml")

					Expect(err).To(HaveOccurred())
				})
			})

			Describe("non-existent files", func() {
				It("fails on loading an non-existent local file", func() {
					_, err := config.ParseArgs("--config", "/does-not-exist.yaml")

					Expect(err).To(HaveOccurred())
				})

				It("fails on loading an non-existent remote file", func() {
					_, err := config.ParseArgs("--config", "http://example.com/doesnotexist.yaml")

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})

	It("can parse an interface filter", func() {
		cfg, err := config.ParseArgs("--interface-filter", "wg[0-9]")
		Expect(err).To(Succeed())

		cfg.Marshal(GinkgoWriter)

		Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
		Expect(cfg.InterfaceFilter("wg1")).To(BeTrue())
		Expect(cfg.InterfaceFilter("et0")).To(BeFalse())
	})

	Describe("use environment variables", func() {
		It("accepts settings via environment variables", func() {
			os.Setenv("CUNICU_EPDISC_ICE_CANDIDATE_TYPES", "srflx")

			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.EndpointDisc.ICE.CandidateTypes).To(ConsistOf(
				icex.CandidateType{CandidateType: ice.CandidateTypeServerReflexive},
			))
		})

		It("accepts multiple settings via environment variables", func() {
			os.Setenv("CUNICU_EPDISC_ICE_CANDIDATE_TYPES", "srflx,relay")

			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.EndpointDisc.ICE.CandidateTypes).To(ConsistOf(
				icex.CandidateType{CandidateType: ice.CandidateTypeServerReflexive},
				icex.CandidateType{CandidateType: ice.CandidateTypeRelay},
			))
		})

		It("environment variables are overwritten by command line arguments", func() {
			cfg, err := config.ParseArgs("--ice-candidate-type", "host")
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.EndpointDisc.ICE.CandidateTypes).To(HaveLen(1))
			Expect(icfg.EndpointDisc.ICE.CandidateTypes).To(ContainElements(
				icex.CandidateType{CandidateType: ice.CandidateTypeHost},
			))
		})
	})

	Describe("use proper default settings", func() {
		var err error
		var cfg *config.Config
		var icfg *config.InterfaceSettings

		BeforeEach(func() {
			cfg, err = config.ParseArgs()
			Expect(err).To(Succeed())

			icfg = &cfg.DefaultInterfaceSettings
		})

		It("should use the standard cunicu signaling backend", func() {
			Expect(cfg.Backends).To(HaveLen(1))
			Expect(cfg.Backends[0].String()).To(Equal("grpc://signal.cunicu.li"))
		})

		It("should accept all interfaces", func() {
			Expect(cfg.InterfaceFilter("wg12345")).To(BeTrue())
		})

		It("should have a default STUN URL", func() {
			Expect(icfg.EndpointDisc.ICE.URLs).To(HaveLen(1))
			Expect(icfg.EndpointDisc.ICE.URLs[0].String()).To(Equal("stun:stun.cunicu.li:3478"))
		})
	})

	Describe("dump", func() {
		var cfg1, cfg2 *config.Config
		var icfg1, icfg2 *config.InterfaceSettings

		BeforeEach(func() {
			var err error

			args := []string{"--ice-network-type", "udp4,udp6", "--url", "stun:stun.cunicu.de", "wg0"}

			cfg1, err = config.ParseArgs(args...)
			Expect(err).To(Succeed())

			buf := &bytes.Buffer{}

			Expect(cfg1.Marshal(buf)).To(Succeed())

			cfg2, err = config.ParseArgs(args...)
			Expect(err).To(Succeed())

			err = cfg2.Koanf.Load(rawbytes.Provider(buf.Bytes()), yaml.Parser())
			Expect(err).To(Succeed())

			err = cfg2.Load()
			Expect(err).To(Succeed())

			icfg1 = &cfg1.DefaultInterfaceSettings
			icfg2 = &cfg2.DefaultInterfaceSettings

		})

		It("have equal WireGuard interface lists", func() {
			Expect(cfg1.Interfaces).To(Equal(cfg2.Interfaces))
		})

		It("have equal ICE network types", func() {
			Expect(icfg1.EndpointDisc.ICE.NetworkTypes).To(Equal(icfg2.EndpointDisc.ICE.NetworkTypes))
		})

		It("have equal ICE URLs", func() {
			Expect(icfg1.EndpointDisc.ICE.URLs).To(Equal(icfg2.EndpointDisc.ICE.URLs))
		})
	})

	It("can parse the example config file", func() {
		cfg, err := config.ParseArgs("--config", "../../etc/cunicu.yaml")
		Expect(err).To(Succeed())

		Expect(cfg.InterfaceSettings("wg-work-laptop").PeerDisc.Community).To(Equal("mysecret-pass"))
	})

	It("throws an error on an invalid config file path", func() {
		_, err := config.ParseArgs("--config", "_:")
		Expect(err).To(MatchError(HavePrefix("failed to load config file: invalid URL")))
	})

	It("throws an error on an invalid config file URL schema", func() {
		_, err := config.ParseArgs("--config", "smb://is-not-supported")
		Expect(err).To(MatchError(And(
			HavePrefix("failed to load config file"),
			HaveSuffix("unsupported URL scheme: smb"),
		)))
	})

	Describe("matching interface configs", func() {
		// TODO
	})

	Describe("runtime", func() {
		It("can set a single setting", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			Expect(cfg.Get("watch_interval")).To(BeNumerically("==", 1*time.Second))
			Expect(cfg.WatchInterval).To(Equal(1 * time.Second))

			Expect(cfg.Set("watch_interval", "100s")).To(Succeed())

			Expect(cfg.Get("watch_interval")).To(Equal("100s"))
			Expect(cfg.WatchInterval).To(Equal(100 * time.Second))
		})

		It("can update multiple settings", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			Expect(cfg.Update(map[string]any{
				"wireguard.listen_port_range.min": 100,
				"wireguard.listen_port_range.max": 200,
			})).To(Succeed())

			Expect(cfg.DefaultInterfaceSettings.WireGuard.ListenPortRange.Min).To(Equal(100))
			Expect(cfg.DefaultInterfaceSettings.WireGuard.ListenPortRange.Max).To(Equal(200))
		})

		It("fails to update multiple settings which are incorrect", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			orig := cfg.DefaultInterfaceSettings.WireGuard.ListenPortRange

			Expect(cfg.Update(map[string]any{
				"wireguard.listen_port_range.min": 200,
				"wireguard.listen_port_range.max": 100,
			})).To(MatchError(
				MatchRegexp(`failed to load config: invalid settings: WireGuard minimal listen port \(\d+\) must be smaller or equal than maximal port \(\d+\)`),
			))

			Expect(cfg.DefaultInterfaceSettings.WireGuard.ListenPortRange).To(Equal(orig), "Failed update has changed settings")
		})

		It("can save runtime settings", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			Expect(cfg.Set("watch_interval", "100s")).To(Succeed())

			buf := &bytes.Buffer{}
			Expect(cfg.MarshalRuntime(buf)).To(Succeed())
			Expect(buf.Bytes()).To(MatchYAML("watch_interval: 100s"))
		})

		It("can register handler for changed settings", Pending, func() {
			// TODO
		})
	})

	Describe("interface overwrites", func() {
		It("single interface", func() {
			cfgFile := mkTempFile(`---
interfaces:
  wg0:
    epdisc:
      ice:
        restart_timeout: 10s
`)

			cfg, err := config.ParseArgs("--config", cfgFile.Name())
			Expect(err).To(Succeed())

			icfg := cfg.InterfaceSettings("wg0")
			Expect(icfg).NotTo(BeNil())

			Expect(cfg.DefaultInterfaceSettings.EndpointDisc.ICE.RestartTimeout).To(Equal(5 * time.Second))
			Expect(icfg.EndpointDisc.ICE.RestartTimeout).To(Equal(10 * time.Second))
			Expect(icfg.Name).To(Equal("wg0"))
			Expect(icfg.Pattern).To(BeEmpty())
		})

		It("single interface and pattern overwrites", func() {
			cfgFile := mkTempFile(`---
interfaces:
  wg0:
    epdisc:
      ice:
        restart_timeout: 10s

  wg-work-*:
    epdisc:
      ice:
        keepalive_interval: 123s
        restart_timeout: 20s

  wg-*:
    epdisc:
      ice:
        restart_timeout: 30s
`)

			cfg, err := config.ParseArgs("--config", cfgFile.Name())
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceOrder).To(Equal([]string{"*", "wg0", "wg-work-*", "wg-*"}))

			icfg1 := cfg.InterfaceSettings("wg-work-seattle")
			Expect(icfg1).NotTo(BeNil())

			icfg2 := cfg.InterfaceSettings("wg-mobile")
			Expect(icfg2).NotTo(BeNil())

			icfg3 := cfg.InterfaceSettings("wg0")
			Expect(icfg3).NotTo(BeNil())

			Expect(icfg1.EndpointDisc.ICE.RestartTimeout).To(Equal(30 * time.Second))
			Expect(icfg1.EndpointDisc.ICE.KeepaliveInterval).To(Equal(123 * time.Second))

			Expect(icfg2.EndpointDisc.ICE.RestartTimeout).To(Equal(30 * time.Second))
			Expect(icfg2.EndpointDisc.ICE.KeepaliveInterval).To(Equal(cfg.DefaultInterfaceSettings.EndpointDisc.ICE.KeepaliveInterval))

			Expect(icfg3.EndpointDisc.ICE.RestartTimeout).To(Equal(10 * time.Second))
		})
	})
})
