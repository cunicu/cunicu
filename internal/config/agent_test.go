package config_test

import (
	"github.com/pion/ice/v2"
	"riasc.eu/wice/internal/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent config", func() {
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
			cfg, err := config.ParseArgs("--ice-interface-filter", "eth\\d+")
			Expect(err).To(Succeed())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())

			Expect(aCfg.InterfaceFilter("eth0")).To(BeTrue())
			Expect(aCfg.InterfaceFilter("wifi0")).To(BeFalse())
		})
	})

	Context("Interface filter with invalid regex", func() {
		It("can parse an interface filter", func() {
			_, err := config.ParseArgs("--ice-interface-filter", "eth(")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("default values", func() {
		It("has sensible default values", func() {
			cfg, err := config.ParseArgs("--ice-interface-filter", "eth\\d+")
			Expect(err).To(Succeed())

			aCfg, err := cfg.AgentConfig()
			Expect(err).To(Succeed())

			Expect(aCfg.Urls).To(HaveLen(1))
			Expect(aCfg.Urls[0].String()).To(Equal(config.DefaultURL))
		})
	})
})
