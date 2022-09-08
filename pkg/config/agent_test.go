package config_test

import (
	"github.com/pion/ice/v2"
	"github.com/stv0g/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent config", func() {
	Context("ICE urls", func() {
		It("can parse ICE urls with credentials", func() {
			cfg, err := config.ParseArgs(
				"--url", "stun:server1",
				"--url", "turn:server2:1234?transport=tcp",
				"--username", "user1",
				"--password", "pass1",
			)
			Expect(err).To(Succeed())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())

			Expect(aCfg.Urls).To(Equal([]*ice.URL{
				{
					Scheme:   ice.SchemeTypeSTUN,
					Host:     "server1",
					Port:     3478,
					Proto:    ice.ProtoTypeUDP,
					Username: "user1",
					Password: "pass1",
				},
				{
					Scheme:   ice.SchemeTypeTURN,
					Host:     "server2",
					Port:     1234,
					Proto:    ice.ProtoTypeTCP,
					Username: "user1",
					Password: "pass1",
				},
			}))
		})
	})

	Context("Candidate types", func() {
		It("can parse multiple candidate types", func() {
			cfg, err := config.ParseArgs(
				"--ice-candidate-type", "host",
				"--ice-candidate-type", "relay",
			)
			Expect(err).To(Succeed())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())
			Expect(aCfg.CandidateTypes).To(ConsistOf(ice.CandidateTypeRelay, ice.CandidateTypeHost))
		})
	})

	Context("Network types", func() {
		It("can parse multiple network types", func() {
			cfg, err := config.ParseArgs(
				"--ice-network-type", "udp4",
				"--ice-network-type", "tcp6",
			)
			Expect(err).To(Succeed())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())
			Expect(aCfg.NetworkTypes).To(ConsistOf(ice.NetworkTypeTCP6, ice.NetworkTypeUDP4))
		})
	})

	Context("Comma-separated network types", func() {
		It("can parse multiple network types", func() {
			cfg, err := config.ParseArgs(
				"--ice-network-type", "udp4,tcp6",
			)
			Expect(err).To(Succeed())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())
			Expect(aCfg.NetworkTypes).To(ConsistOf(ice.NetworkTypeTCP6, ice.NetworkTypeUDP4))
		})
	})

	Context("Interface filter", func() {
		It("can parse an interface filter", func() {
			cfg, err := config.ParseArgs("--wg-interface-filter", "wg\\d+")
			Expect(err).To(Succeed())

			Expect(cfg.WireGuard.InterfaceFilter.MatchString("wg0")).To(BeTrue())
			Expect(cfg.WireGuard.InterfaceFilter.MatchString("et0")).To(BeFalse())
		})
	})

	Context("Interface filter with invalid regex", func() {
		It("can parse an interface filter", func() {
			_, err := config.ParseArgs("--wg-interface-filter", "eth(")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("default values", func() {
		It("has default values", func() {
			cfg, err := config.ParseArgs()
			Expect(err).To(Succeed())

			Expect(cfg.WireGuard.InterfaceFilter.MatchString("wg1234")).To(BeTrue())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())

			Expect(aCfg.Urls).To(HaveLen(1))
			Expect(aCfg.Urls[0].String()).To(Equal(config.DefaultURL))
		})
	})
})
