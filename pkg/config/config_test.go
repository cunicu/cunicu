// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/onsi/gomega/ghttp"
	"github.com/pion/ice/v2"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Context("config", func() {
	mkTempFile := func(contents string) *os.File {
		dir := GinkgoT().TempDir()
		fn := filepath.Join(dir, "cunicu.yaml")

		file, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0o600)
		Expect(err).To(Succeed())

		defer file.Close()

		_, err = file.Write([]byte(contents))
		Expect(err).To(Succeed())

		return file
	}

	Describe("parse command line arguments", func() {
		It("can parse a boolean argument like --wg-userspace", func() {
			cfg, err := config.ParseArgs("--wg-userspace")
			Expect(err).To(Succeed())

			Expect(cfg.DefaultInterfaceSettings.UserSpace).To(BeTrue())
		})

		It("can parse multiple backends", func() {
			cfg, err := config.ParseArgs("--backend", "grpc", "--backend", "inprocess")

			Expect(err).To(Succeed())
			Expect(cfg.Backends).To(HaveLen(2))
			Expect(cfg.Backends[0].Scheme).To(Equal("grpc"))
			Expect(cfg.Backends[1].Scheme).To(Equal("inprocess"))
		})

		It("parse an interface list", func() {
			cfg, err := config.ParseArgs("wg0", "wg1")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg2")).To(BeFalse())
		})

		It("parse an interface list with patterns", func() {
			cfg, err := config.ParseArgs("wg0", "wg1", "wg-work-*")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg-work-0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg2")).To(BeFalse())
			Expect(cfg.InterfaceSettings("wg2")).To(BeNil())
		})

		It("fails on invalid arguments", func() {
			_, err := config.ParseArgs("--wrong")

			Expect(err).To(MatchError("failed to parse command line flags: unknown flag: --wrong"))
		})

		It("fails on invalid arguments values", func() {
			_, err := config.ParseArgs("--ice-url", ":_")

			Expect(err).To(MatchError(HaveSuffix("missing protocol scheme")))
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
					BeforeEach(func() {
						// Move config file into XDG config directory
						configDir := filepath.Dir(cfgFile.Name())

						err := os.Setenv("CUNICU_CONFIG_DIR", configDir)
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
						ghttp.VerifyRequest("HEAD", "/cunicu.yaml"),
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
					// shut down the server between tests
					server.Close()
				})

				It("fails on loading an non-existent remote file", func() {
					_, err := config.ParseArgs("--config", "http://example.com/doesnotexist.yaml")

					Expect(err).To(HaveOccurred())
				})
			})

			Describe("non-existing files", func() {
				It("fails on loading an non-existing local file paths", func() {
					var errPattern string
					if runtime.GOOS == "windows" {
						errPattern = `The system cannot find the (path|file) specified.$`
					} else {
						errPattern = `no such file or directory$`
					}

					_, err := config.ParseArgs("--config", "/does-not-exist.yaml")

					Expect(err).To(MatchError(MatchRegexp(errPattern)))
				})

				It("fails on loading an non-existing remote file paths", func() {
					_, err := config.ParseArgs("--config", "https://domain.invalid/config.yaml")

					Expect(err).To(MatchError(MatchRegexp(`^failed to load config: failed to fetch https://domain\.invalid/config\.yaml`)))
				})
			})
		})
	})

	Describe("use environment variables", func() {
		It("accepts settings via environment variables", func() {
			os.Setenv("CUNICU_ICE_CANDIDATE_TYPES", "srflx")

			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.ICE.CandidateTypes).To(ConsistOf(
				icex.CandidateType{
					CandidateType: ice.CandidateTypeServerReflexive,
				},
			))
		})

		It("accepts multiple settings via environment variables", func() {
			os.Setenv("CUNICU_ICE_CANDIDATE_TYPES", "srflx,relay")

			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.ICE.CandidateTypes).To(ConsistOf(
				icex.CandidateType{CandidateType: ice.CandidateTypeServerReflexive},
				icex.CandidateType{CandidateType: ice.CandidateTypeRelay},
			))
		})

		It("environment variables are overwritten by command line arguments", func() {
			os.Setenv("CUNICU_ICE_CANDIDATE_TYPES", "srflx,relay")

			cfg, err := config.ParseArgs("--ice-candidate-type", "host")
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.ICE.CandidateTypes).To(HaveLen(1))
			Expect(icfg.ICE.CandidateTypes).To(ContainElements(
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
			Expect(cfg.Backends[0].String()).To(Equal("grpc://signal.cunicu.li:443"))
		})

		It("should have a default STUN URL", func() {
			Expect(icfg.ICE.URLs).To(HaveLen(1))
			Expect(icfg.ICE.URLs[0].String()).To(Equal("grpc://relay.cunicu.li:443"))
		})
	})

	Describe("dump", func() {
		var cfg1, cfg2 *config.Config
		var icfg1, icfg2 *config.InterfaceSettings

		BeforeEach(func() {
			var err error

			args := []string{"--ice-network-type", "udp4,udp6", "--ice-url", "stun:stun.cunicu.de", "wg0"}

			cfg1, err = config.ParseArgs(args...)
			Expect(err).To(Succeed())

			buf := &bytes.Buffer{}

			Expect(cfg1.Marshal(buf)).To(Succeed())

			cfg2, err = config.ParseArgs(args...)
			Expect(err).To(Succeed())

			err = cfg2.Koanf.Load(rawbytes.Provider(buf.Bytes()), yaml.Parser())
			Expect(err).To(Succeed())

			err = cfg2.Init(nil)
			Expect(err).To(Succeed())

			icfg1 = &cfg1.DefaultInterfaceSettings
			icfg2 = &cfg2.DefaultInterfaceSettings
		})

		It("have equal WireGuard interface lists", func() {
			Expect(cfg1.Interfaces).To(Equal(cfg2.Interfaces))
		})

		It("have equal ICE network types", func() {
			Expect(icfg1.ICE.NetworkTypes).To(Equal(icfg2.ICE.NetworkTypes))
		})

		It("have equal ICE URLs", func() {
			Expect(icfg1.ICE.URLs).To(Equal(icfg2.ICE.URLs))
		})
	})

	Context("allow insecure configs", func() {
		BeforeEach(func() {
			os.Setenv("CUNICU_CONFIG_ALLOW_INSECURE", "true")
		})

		AfterEach(func() {
			os.Unsetenv("CUNICU_CONFIG_ALLOW_INSECURE")
		})

		It("can parse the example config file", func() {
			cfg, err := config.ParseArgs("--config", "../../etc/cunicu.advanced.yaml")

			Expect(err).To(Succeed())

			Expect(cfg.Files).To(Equal([]string{"../../etc/cunicu.advanced.yaml"}))
			Expect(cfg.InterfaceOrder).To(Equal([]string{"wg0", "wg1", "wg2", "wg-work-*", "wg-work-external-*"}))
			Expect(cfg.InterfaceSettings("wg-work-laptop").Community).To(BeEquivalentTo(crypto.GenerateKeyFromPassword("mysecret-pass")))
			Expect(cfg.DefaultInterfaceSettings.Hooks).To(HaveLen(2))

			h := cfg.DefaultInterfaceSettings.Hooks[0]
			hh, ok := h.(*config.ExecHookSetting)
			Expect(ok).To(BeTrue(), "Found invalid hook %+#v", hh)
		})
	})

	It("throws an error on an invalid config file path", func() {
		_, err := config.ParseArgs("--config", "_:")
		Expect(err).To(MatchError(HavePrefix("ignoring config file with invalid name")))
	})

	It("throws an error on an invalid config file URL schema", func() {
		_, err := config.ParseArgs("--config", "smb://is-not-supported")
		Expect(err).To(MatchError("unsupported scheme 'smb' for config file"))
	})

	Describe("runtime", func() {
		It("can update multiple settings", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			_, err = cfg.Update(map[string]any{
				"watch_interval":        100 * time.Second,
				"listen_port_range.min": 100,
				"listen_port_range.max": 200,
			})
			Expect(err).To(Succeed())

			Expect(cfg.WatchInterval).To(Equal(100 * time.Second))
			Expect(cfg.DefaultInterfaceSettings.ListenPortRange.Min).To(Equal(100))
			Expect(cfg.DefaultInterfaceSettings.ListenPortRange.Max).To(Equal(200))
		})

		It("fails to update multiple settings which are incorrect", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			orig := cfg.DefaultInterfaceSettings.ListenPortRange

			_, err = cfg.Update(map[string]any{
				"listen_port_range.min": 200,
				"listen_port_range.max": 100,
			})
			Expect(err).To(MatchError(
				MatchRegexp(`invalid settings: WireGuard minimal listen port \(\d+\) must be smaller or equal than maximal port \(\d+\)`),
			))

			Expect(cfg.DefaultInterfaceSettings.ListenPortRange).To(Equal(orig), "Failed update has changed settings")
		})

		It("can save runtime settings", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			_, err = cfg.Update(map[string]any{
				"watch_interval": "100s",
			})
			Expect(err).To(Succeed())

			buf := &bytes.Buffer{}
			Expect(cfg.MarshalRuntime(buf)).To(Succeed())
			Expect(buf.Bytes()).To(MatchYAML("watch_interval: 100s"))
		})
	})

	Describe("reload", func() {
	})

	Describe("interface overwrites", func() {
		It("should accept all interfaces", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceOrder).To(Equal([]string{"*"}))

			icfg := cfg.InterfaceSettings("wg12345")
			Expect(icfg).NotTo(BeNil())

			Expect(*icfg).To(Equal(cfg.DefaultInterfaceSettings))
		})

		It("single interface as argument", func() {
			cfg, err := config.ParseArgs("wg0")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceOrder).To(Equal([]string{"wg0"}))

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeFalse())

			icfg := cfg.InterfaceSettings("wg0")
			Expect(icfg).NotTo(BeNil())

			Expect(*icfg).To(Equal(cfg.DefaultInterfaceSettings))
		})

		It("single interface pattern as argument", func() {
			cfg, err := config.ParseArgs("wg*")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceOrder).To(Equal([]string{"wg*"}))

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("eth0")).To(BeFalse())

			icfg := cfg.InterfaceSettings("wg0")
			Expect(icfg).NotTo(BeNil())

			Expect(*icfg).To(Equal(cfg.DefaultInterfaceSettings))
		})

		It("single interface", func() {
			cfgFile := mkTempFile(`---
ice:
  restart_timeout: 5s
  disconnected_timeout: 22s

interfaces:
  wg0:
    ice:
      restart_timeout: 10s
`)

			cfg, err := config.ParseArgs("--config", cfgFile.Name())
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeFalse())

			Expect(cfg.InterfaceOrder).To(Equal([]string{"wg0"}))

			icfg := cfg.InterfaceSettings("wg0")
			Expect(icfg).NotTo(BeNil())

			Expect(icfg.ICE.RestartTimeout).To(Equal(10 * time.Second))
			Expect(icfg.ICE.DisconnectedTimeout).To(Equal(22 * time.Second))

			Expect(cfg.DefaultInterfaceSettings.ICE.RestartTimeout).To(Equal(5 * time.Second))
		})

		It("two interface names and two patterns", func() {
			cfgFile := mkTempFile(`---
ice:
  keepalive_interval: 7s

interfaces:
  wg0:
    ice:
      restart_timeout: 10s

  wg-work-*:
    ice:
      keepalive_interval: 123s
      restart_timeout: 20s

  wg-*:
    ice:
      restart_timeout: 30s
`)

			cfg, err := config.ParseArgs("--config", cfgFile.Name(), "wg1")
			Expect(err).To(Succeed())

			Expect(cfg.InterfaceOrder).To(Equal([]string{"wg1", "wg0", "wg-work-*", "wg-*"}))

			Expect(cfg.InterfaceFilter("wg0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg1")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg-work-0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("wg-0")).To(BeTrue())
			Expect(cfg.InterfaceFilter("eth0")).To(BeFalse())

			icfg1 := cfg.InterfaceSettings("wg-work-seattle")
			Expect(icfg1).NotTo(BeNil())

			Expect(icfg1.ICE.RestartTimeout).To(Equal(30 * time.Second))
			Expect(icfg1.ICE.KeepaliveInterval).To(Equal(123 * time.Second))

			icfg2 := cfg.InterfaceSettings("wg-mobile")
			Expect(icfg2).NotTo(BeNil())

			Expect(icfg2.ICE.RestartTimeout).To(Equal(30 * time.Second))
			Expect(icfg2.ICE.KeepaliveInterval).To(Equal(7 * time.Second))

			icfg3 := cfg.InterfaceSettings("wg0")
			Expect(icfg3).NotTo(BeNil())

			Expect(icfg3.ICE.RestartTimeout).To(Equal(10 * time.Second))
			Expect(icfg3.ICE.KeepaliveInterval).To(Equal(7 * time.Second))

			icfg4 := cfg.InterfaceSettings("wg1")
			Expect(icfg4).NotTo(BeNil())

			Expect(*icfg4).To(Equal(cfg.DefaultInterfaceSettings))
		})
	})
})
