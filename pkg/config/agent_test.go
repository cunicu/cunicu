package config_test

import (
	"context"

	"github.com/pion/ice/v2"
	"github.com/stv0g/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Agent config", func() {
	DescribeTable("can parse ICE urls with credentials",
		func(urlStr string, exp any) {
			cfg, err := config.ParseArgs(
				"--url", urlStr,
				"--username", "user1",
				"--password", "pass1",
			)
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			aCfg, err := icfg.AgentConfig(context.Background(), nil)

			switch exp := exp.(type) {
			case string:
				Expect(err).To(MatchError(exp))
			case *ice.URL:
				Expect(err).To(Succeed())
				Expect(aCfg.Urls).To(ContainElements(exp))
			}
		},
		Entry("url1", "stun:server1", &ice.URL{
			Scheme:   ice.SchemeTypeSTUN,
			Host:     "server1",
			Port:     3478,
			Proto:    ice.ProtoTypeUDP,
			Username: "user1",
			Password: "pass1",
		}),
		Entry("url2", "turn:server2:1234?transport=tcp", &ice.URL{
			Scheme:   ice.SchemeTypeTURN,
			Host:     "server2",
			Port:     1234,
			Proto:    ice.ProtoTypeTCP,
			Username: "user1",
			Password: "pass1",
		}),
		Entry("url3", "turn:user3:pass3@server3:1234?transport=tcp", &ice.URL{
			Scheme:   ice.SchemeTypeTURN,
			Host:     "server3",
			Port:     1234,
			Proto:    ice.ProtoTypeTCP,
			Username: "user3",
			Password: "pass3",
		}),
		Entry("url4", "turn:user3@server3:1234?transport=tcp", "failed to gather ICE URLs: invalid user / password"),
		Entry("url5", "http://bla.0l.de", "failed to gather ICE URLs: invalid ICE URL scheme: http"),
		Entry("url6", "stun:stun.cunicu.li?transport=tcp", "failed to gather ICE URLs: failed to parse STUN/TURN URL 'stun:stun.cunicu.li?transport=tcp': queries not supported in stun address"),
	)

	It("can parse relay api ICE urls", Pending, func() {
		// TODO
	})

	It("can parse multiple candidate types", func() {
		cfg, err := config.ParseArgs(
			"--ice-candidate-type", "host",
			"--ice-candidate-type", "relay",
		)
		Expect(err).To(Succeed())

		icfg := cfg.DefaultInterfaceSettings

		aCfg, err := icfg.AgentConfig(context.Background(), nil)
		Expect(err).To(Succeed())
		Expect(aCfg.CandidateTypes).To(ConsistOf(ice.CandidateTypeRelay, ice.CandidateTypeHost))
	})

	It("can parse multiple network types when passed as individual command line arguments", func() {
		cfg, err := config.ParseArgs(
			"--ice-network-type", "udp4",
			"--ice-network-type", "tcp6",
		)
		Expect(err).To(Succeed())

		icfg := cfg.DefaultInterfaceSettings

		aCfg, err := icfg.AgentConfig(context.Background(), nil)
		Expect(err).To(Succeed())
		Expect(aCfg.NetworkTypes).To(ConsistOf(ice.NetworkTypeTCP6, ice.NetworkTypeUDP4))
	})

	It("can parse multiple network types when passed as comma-separated value", func() {
		cfg, err := config.ParseArgs("--ice-network-type", "udp4,tcp6")
		Expect(err).To(Succeed())

		icfg := cfg.DefaultInterfaceSettings

		aCfg, err := icfg.AgentConfig(context.Background(), nil)
		Expect(err).To(Succeed())
		Expect(aCfg.NetworkTypes).To(ConsistOf(ice.NetworkTypeTCP6, ice.NetworkTypeUDP4))
	})

	It("has proper default values", func() {
		cfg, err := config.ParseArgs()
		Expect(err).To(Succeed())

		icfg := cfg.DefaultInterfaceSettings

		aCfg, err := icfg.AgentConfig(context.Background(), nil)
		Expect(err).To(Succeed())

		Expect(aCfg.InterfaceFilter("wg1")).To(BeTrue())

		Expect(aCfg.Urls).To(HaveLen(1))
		Expect(aCfg.Urls[0].String()).To(Equal(config.DefaultICEURLs[0].String()))
	})
})
