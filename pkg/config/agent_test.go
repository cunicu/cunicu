// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/pion/stun/v2"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"cunicu.li/cunicu/pkg/config"
	"cunicu.li/cunicu/pkg/crypto"
	grpcx "cunicu.li/cunicu/pkg/signaling/grpc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func parseRaw(raw string) (*config.Config, error) {
	flags := pflag.NewFlagSet("", pflag.ContinueOnError)
	cfg := config.New(flags)

	if err := cfg.AddProvider(rawbytes.Provider([]byte(raw))); err != nil {
		return nil, err
	}

	if _, err := cfg.ReloadAllSources(); err != nil {
		return nil, err
	}

	return cfg, nil
}

var _ = Describe("Agent config", func() {
	var err error
	var pk crypto.Key

	BeforeEach(func() {
		pk, err = crypto.GenerateKey()
		Expect(err).To(Succeed())
	})

	DescribeTable("can parse ICE urls with credentials",
		func(url, username, password string, exp any) {
			iceCfgStr := fmt.Sprintf("urls: [ '%s' ], candidate_types: [ srflx ]", url)

			if username != "" {
				iceCfgStr += fmt.Sprintf(", username: '%s'", username)
			}

			if password != "" {
				iceCfgStr += fmt.Sprintf(", password: '%s'", password)
			}

			cfgStr := fmt.Sprintf("ice: { %s }", iceCfgStr)
			cfg, err := parseRaw(cfgStr)
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			aCfg, err := icfg.AgentConfig(context.Background(), &pk)

			switch exp := exp.(type) {
			case string:
				Expect(err).To(MatchError(exp))
			case *stun.URI:
				Expect(err).To(Succeed())
				Expect(aCfg.Urls).To(HaveLen(1))
				Expect(aCfg.Urls).To(ContainElements(exp))
			}
		},
		Entry("url1", "stun:server1", "user1", "pass1", &stun.URI{
			Scheme:   stun.SchemeTypeSTUN,
			Host:     "server1",
			Port:     3478,
			Proto:    stun.ProtoTypeUDP,
			Username: "user1",
			Password: "pass1",
		}),
		Entry("url2", "turn:server2:1234?transport=tcp", "user1", "pass1", &stun.URI{
			Scheme:   stun.SchemeTypeTURN,
			Host:     "server2",
			Port:     1234,
			Proto:    stun.ProtoTypeTCP,
			Username: "user1",
			Password: "pass1",
		}),
		Entry("url3", "turn:user3:pass3@server3:1234?transport=tcp", "", "pass3", &stun.URI{
			Scheme:   stun.SchemeTypeTURN,
			Host:     "server3",
			Port:     1234,
			Proto:    stun.ProtoTypeTCP,
			Username: "user3",
			Password: "pass3",
		}),
		Entry("url3", "http://bla.0l.de", "", "pass1", "failed to gather ICE URLs: invalid ICE URL scheme: http"),
		Entry("url4", "stun:stun.cunicu.li?transport=tcp", "", "", "failed to gather ICE URLs: failed to parse STUN/TURN URL 'stun:stun.cunicu.li?transport=tcp': queries not supported in stun address"),
	)

	Context("can get ICE URLs from relay API", func() {
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

			relays := []grpcx.RelayInfo{stunRelay, turnRelay}

			svr, err = grpcx.NewRelayAPIServer(relays, grpc.Creds(insecure.NewCredentials()))
			Expect(err).To(Succeed())

			l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			Expect(err).To(Succeed())

			go func() {
				err := svr.Serve(l)
				Expect(err).To(Succeed())
			}()
		})

		AfterEach(func() {
			err := svr.Close()
			Expect(err).To(Succeed())
		})

		It("can get list of relays", func() {
			cfgStr := fmt.Sprintf("ice: { urls: [ 'grpc://localhost:%d?insecure=true' ] }", port)
			cfg, err := parseRaw(cfgStr)
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
				case stun.SchemeTypeSTUN:
					Expect(u.String()).To(Equal(stunRelay.URL))
					Expect(u.Username).To(BeEmpty())
					Expect(u.Password).To(BeEmpty())

				case stun.SchemeTypeTURN:
					Expect(u.String()).To(Equal(turnRelay.URL))

					user, pass, exp := turnRelay.GetCredentials(pk.String())

					Expect(strings.Split(user, ":")).To(Equal([]string{
						fmt.Sprint(exp.Unix()),
						pk.String(),
					}))
					Expect(u.Password).To(Equal(pass))

				case stun.SchemeTypeSTUNS, stun.SchemeTypeTURNS, stun.SchemeTypeUnknown:
				}
			}
		})
	})

	It("can parse multiple backend URLs when passed as individual command line arguments", func() {
		cfg, err := parseArgs(
			"--backend", "grpc://server1",
			"--backend", "grpc://server2",
		)
		Expect(err).To(Succeed())

		Expect(cfg.Backends).To(HaveLen(2))
		Expect(cfg.Backends[0].Host).To(Equal("server1"))
		Expect(cfg.Backends[1].Host).To(Equal("server2"))
	})

	It("can parse multiple backend URLs when passed as comma-separated command line arguments", func() {
		cfg, err := parseArgs(
			"--backend", "grpc://server1,grpc://server2",
		)
		Expect(err).To(Succeed())

		Expect(cfg.Backends).To(HaveLen(2))
	})
})
