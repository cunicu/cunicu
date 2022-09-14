package config_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pion/ice/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	grpcx "github.com/stv0g/cunicu/pkg/signaling/grpc"
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
		Entry("url3", "http://bla.0l.de", "failed to gather ICE URLs: invalid ICE URL scheme: http"),
		Entry("url4", "stun:stun.cunicu.li?transport=tcp", "failed to gather ICE URLs: failed to parse STUN/TURN URL 'stun:stun.cunicu.li?transport=tcp': queries not supported in stun address"),
	)

	Context("can getch ICE urls from relay API", func() {
		var err error
		var svr *grpcx.RelayAPIServer
		var stunRelay, turnRelay grpcx.RelayInfo
		var port int

		BeforeEach(func() {
			port = 1234

			stunRelay = grpcx.RelayInfo{
				URL:   "stun:cunicu.li:3478",
				Realm: "cunicu.li",
				// STUN servers need no authentication => no secret and TTL
			}

			turnRelay = grpcx.RelayInfo{
				URL:    "turn:cunicu.li:3478?transport=udp",
				Realm:  "cunicu.li",
				TTL:    1 * time.Hour,
				Secret: "my-very-secret-secret",
			}

			relays := []string{stunRelay.URL, turnRelay.URL}

			svr, err = grpcx.NewRelayAPIServer(relays, grpc.Creds(insecure.NewCredentials()))
			Expect(err).To(Succeed())

			l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			Expect(err).To(Succeed())

			go svr.Serve(l)
		})

		AfterEach(func() {
			err := svr.Close()
			Expect(err).To(Succeed())
		})

		It("can get list of relays", func() {
			cfg, err := config.ParseArgs("--url", fmt.Sprintf("grpc://localhost:%d?insecure=true", port))
			Expect(err).To(Succeed())

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			sk, err := crypto.GenerateKey()
			Expect(err).To(Succeed())

			pk := sk.PublicKey()

			urls, err := cfg.DefaultInterfaceSettings.AgentURLs(ctx, &pk)
			Expect(err).To(Succeed())

			Expect(urls).To(HaveLen(2))
			for _, u := range urls {
				switch u.Scheme {
				case ice.SchemeTypeSTUN:
					Expect(u.String()).To(Equal(stunRelay.URL))
					Expect(u.Username).To(BeEmpty())
					Expect(u.Password).To(BeEmpty())

				case ice.SchemeTypeTURN:
					Expect(u.String()).To(Equal(turnRelay.URL))

					user, pass, exp := turnRelay.GetCredentials(pk.String())

					Expect(strings.Split(user, ":")).To(Equal([]string{
						fmt.Sprint(exp.Unix()),
						pk.String(),
					}))
					Expect(u.Password).To(Equal(pass))
				}
			}
		})
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
