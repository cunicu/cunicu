package config_test

import (
	"github.com/pion/ice/v2"
	"riasc.eu/wice/pkg/config"

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

	It("Some more arguments", func() {
		cfg, err := config.ParseArgs("--ice-mdns", "--ice-nat-1to1-ip=1.2.3.4,4.5.6.7", "--ice-nat-1to1-ip", "10.10.10.10")
		Expect(err).To(Succeed())

		aCfg, err := cfg.AgentConfig()
		Expect(err).To(Succeed())

		Expect(aCfg.MulticastDNSMode).To(Equal(ice.MulticastDNSModeQueryAndGather))
		Expect(aCfg.NAT1To1IPs).To(Equal([]string{"1.2.3.4", "4.5.6.7", "10.10.10.10"}))
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
