package config_test

import (
	"bytes"
	"net/http"
	"os"
	"testing"
	"time"

	"riasc.eu/wice/pkg/config"
	icex "riasc.eu/wice/pkg/ice"
	"riasc.eu/wice/pkg/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pion/ice/v2"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = test.SetupLogging()

var _ = Describe("parse command line arguments", func() {
	It("can parse a boolean argument like wg-userspace", func() {
		c, err := config.ParseArgs("--wg-userspace")

		Expect(err).To(Succeed())
		Expect(c.WireGuard.Userspace).To(BeTrue())
	})

	It("can parse multiple backends", func() {
		c, err := config.ParseArgs("--backend", "k8s", "--backend", "p2p")

		Expect(err).To(Succeed())
		Expect(c.Backends).To(HaveLen(2))
		Expect(c.Backends[0].Scheme).To(Equal("k8s"))
		Expect(c.Backends[1].Scheme).To(Equal("p2p"))
	})

	It("can parse a duration value", func() {
		c, err := config.ParseArgs("--ice-restart-timeout", "10s")

		Expect(err).To(Succeed())
		Expect(c.EndpointDisc.ICE.RestartTimeout).To(Equal(10 * time.Second))
	})

	It("parse an interface list", func() {
		c, err := config.ParseArgs("wg0", "wg1")

		Expect(err).To(Succeed())
		Expect(c.WireGuard.Interfaces).To(ConsistOf("wg0", "wg1"))
	})

	It("fails on invalid arguments", func() {
		_, err := config.ParseArgs("--wrong")

		Expect(err).To(HaveOccurred())
	})

	It("should not load anything from domains without ɯice auto-configuration", func() {
		_, err := config.ParseArgs("-A", "google.com")

		Expect(err).To(
			And(
				MatchError(ContainSubstring("DNS auto-configuration failed")),
				Or(
					MatchError(ContainSubstring("no such host")),
					MatchError(ContainSubstring("i/o timeout")),
				),
			),
		)
	})

	It("should fail when passed an non-existent domain name", func() {
		// RFC6761 defines that "invalid" is a special domain name to always be invalid
		_, err := config.ParseArgs("-A", "invalid")

		Expect(err).To(HaveOccurred())
	})

	Describe("parse configuration files", func() {

		Context("with a local file", func() {
			var cfgFile *os.File

			Context("file with explicit path", func() {
				BeforeEach(func() {
					var err error
					cfgFile, err = os.CreateTemp("", "wice-*.yaml")
					Expect(err).To(Succeed())
				})

				It("can read a single valid local file", func() {
					Expect(cfgFile.WriteString("watch_interval: 1337s\n")).To(BeNumerically(">", 0))
					Expect(cfgFile.Close()).To(Succeed())

					c, err := config.ParseArgs("--config", cfgFile.Name())

					Expect(err).To(Succeed())
					Expect(c.WatchInterval).To(Equal(1337 * time.Second))
				})

				Specify("that command line arguments take precedence over settings provided by configuration files", func() {
					Expect(cfgFile.WriteString("watch_interval: 1337s\n")).To(BeNumerically(">", 0))
					Expect(cfgFile.Close()).To(Succeed())

					c, err := config.ParseArgs("--config", cfgFile.Name(), "--watch-interval", "1m")
					Expect(err).To(Succeed())
					Expect(c.WatchInterval).To(Equal(time.Minute))
				})
			})

			Context("in search path", func() {
				BeforeEach(func() {
					var err error
					cfgFile, err = os.OpenFile("wice.json", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
					Expect(err).To(Succeed())
				})

				It("can read a single valid local file", func() {
					Expect(cfgFile.WriteString(`{ "watch_interval": "1337s" }`)).To(BeNumerically(">", 0))
					Expect(cfgFile.Close()).To(Succeed())

					c, err := config.ParseArgs("--config", cfgFile.Name())

					Expect(err).To(Succeed())
					Expect(c.WatchInterval).To(Equal(1337 * time.Second))
				})
			})

			AfterEach(func() {
				os.RemoveAll(cfgFile.Name())
			})
		})

		Context("with a remote URL", func() {
			var server *ghttp.Server

			BeforeEach(func() {
				server = ghttp.NewServer()
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/wice.yaml"),
						ghttp.RespondWith(http.StatusOK,
							"community: this-is-a-test\n",
							http.Header{
								"Content-type": []string{"text/yaml"},
							}),
					),
				)
			})

			It("can fetch a valid remote configuration file", func() {
				cfg, err := config.ParseArgs("--config", server.URL()+"/wice.yaml")

				Expect(err).To(Succeed())
				Expect(cfg.Community).To(Equal("this-is-a-test"))
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

var _ = Describe("use environment variables", func() {
	BeforeEach(func() {
		os.Setenv("WICE_ENDPOINT_DISC_ICE_CANDIDATE_TYPES", "srflx,relay")
	})

	It("accepts settings via environment variables", func() {
		c, err := config.ParseArgs()
		Expect(err).To(Succeed())

		Expect(c.EndpointDisc.ICE.CandidateTypes).To(ConsistOf(
			icex.CandidateType{CandidateType: ice.CandidateTypeServerReflexive},
			icex.CandidateType{CandidateType: ice.CandidateTypeRelay},
		))
	})

	It("environment variables are overwritten by command line arguments", func() {
		c, err := config.ParseArgs("--ice-candidate-type", "host")
		Expect(err).To(Succeed())

		Expect(c.EndpointDisc.ICE.CandidateTypes).To(ConsistOf(
			icex.CandidateType{CandidateType: ice.CandidateTypeHost},
		))
	})
})

var _ = Describe("use proper default options", func() {
	var c *config.Config

	BeforeEach(func() {
		var err error
		c, err = config.ParseArgs()

		Expect(err).To(Succeed())
	})

	It("should have a default STUN URL", func() {
		Expect(c.EndpointDisc.ICE.URLs).To(HaveLen(1))
		Expect(c.EndpointDisc.ICE.URLs).To(ContainElement(HaveField("Host", "l.google.com")))
	})
})

var _ = Describe("dump", func() {
	var c1, c2 *config.Config

	BeforeEach(func() {
		var err error
		c1, err = config.ParseArgs("--ice-network-type", "udp4,udp6", "--url", "stun:0l.de", "wg0")
		Expect(err).To(Succeed())

		buf := &bytes.Buffer{}

		Expect(c1.Dump(buf)).To(Succeed())

		c2 = config.NewConfig(nil)
		c2.SetConfigType("yaml")

		Expect(c2.MergeConfig(buf)).To(Succeed())
		Expect(c2.Load()).To(Succeed())
	})

	It("have equal WireGuard interface lists", func() {
		Expect(c1.WireGuard.Interfaces).To(Equal(c2.WireGuard.Interfaces))
	})

	It("have equal ICE network types", func() {
		Expect(c1.EndpointDisc.ICE.NetworkTypes).To(Equal(c2.EndpointDisc.ICE.NetworkTypes))
	})

	It("have equal ICE URLs", func() {
		Expect(c1.EndpointDisc.ICE.URLs).To(Equal(c2.EndpointDisc.ICE.URLs))
	})
})

var _ = It("can parse the example config file", func() {
	_, err := config.ParseArgs("--config", "../../etc/wice.yaml")
	Expect(err).To(Succeed())
})
