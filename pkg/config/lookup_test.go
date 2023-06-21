// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/foxcpp/go-mockdns"
	"github.com/onsi/gomega/ghttp"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lookup", func() {
	It("should not load anything from domains without auto-configuration", func() {
		_, err := config.ParseArgs("-D", "google.com")

		Expect(err).To(
			And(
				MatchError(ContainSubstring("failed to load config")),
				Or(
					MatchError(ContainSubstring("no such host")),
					MatchError(ContainSubstring("i/o timeout")),
					MatchError(ContainSubstring("DNS name does not exist")), // raised by Windows
				),
			),
		)
	})

	It("should fail when passed an non-existent domain name", func() {
		// RFC6761 defines that "invalid" is a special domain name to always be invalid
		_, err := config.ParseArgs("-D", "invalid")

		Expect(err).To(HaveOccurred())
	})

	Context("mockup dns", func() {
		var dnsSrv *mockdns.Server
		var webSrv *ghttp.Server

		cfgPath := "/cunicu"

		BeforeEach(func() {
			var err error

			webSrv = ghttp.NewServer()
			webSrv.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", cfgPath),
					ghttp.VerifyHeader(http.Header{
						"User-agent": []string{buildinfo.UserAgent()},
					}),
					ghttp.RespondWith(http.StatusOK, "{ interfaces: { wg-test: { } } }",
						http.Header{
							"Content-type": []string{"text/yaml"},
						}),
				),
			)

			dnsSrv, err = mockdns.NewServerWithLogger(map[string]mockdns.Zone{
				"example.com.": {
					A: []string{"1.2.3.4"},
					TXT: []string{
						"cunicu-backend=p2p",
						"cunicu-backend=grpc://example.com:8080",
						"cunicu-community=my-community-password",
						"cunicu-ice-username=user1",
						"cunicu-ice-password=pass1",
						fmt.Sprintf("cunicu-config=%s%s", webSrv.URL(), cfgPath),
					},
				},
				"_stun._udp.example.com.": {
					SRV: []net.SRV{
						{
							Target:   "stun.example.com.",
							Port:     3478,
							Priority: 10,
							Weight:   0,
						},
					},
				},
				"_stuns._tcp.example.com.": {
					SRV: []net.SRV{
						{
							Target:   "stun.example.com.",
							Port:     3478,
							Priority: 10,
							Weight:   0,
						},
					},
				},
				"_turn._udp.example.com.": {
					SRV: []net.SRV{
						{
							Target:   "turn.example.com.",
							Port:     3478,
							Priority: 10,
							Weight:   0,
						},
					},
				},
				"_turn._tcp.example.com.": {
					SRV: []net.SRV{
						{
							Target:   "turn.example.com.",
							Port:     3478,
							Priority: 10,
							Weight:   0,
						},
					},
				},
				"_turns._tcp.example.com.": {
					SRV: []net.SRV{
						{
							Target:   "turn.example.com.",
							Port:     5349,
							Priority: 10,
							Weight:   0,
						},
					},
				},
			}, GinkgoWriter, false)
			Expect(err).To(Succeed())

			dnsSrv.PatchNet(net.DefaultResolver)
		})

		AfterEach(func() {
			mockdns.UnpatchNet(net.DefaultResolver)

			err := dnsSrv.Close()
			Expect(err).To(Succeed())
		})

		It("check mock dns server", func() {
			addr, err := net.ResolveIPAddr("ip", "example.com")
			Expect(err).To(Succeed())
			Expect(addr.IP.String()).To(Equal("1.2.3.4"))
		})

		It("can get SOA serial", func() {
			p := config.NewLookupProvider("example.com")

			v := p.Version()
			Expect(v).NotTo(BeNil())

			s, ok := v.(int)
			Expect(ok).To(BeTrue())

			Expect(s).To(Equal(1))
		})

		It("can do DNS auto configuration", func() {
			cfg, err := config.ParseArgs("--domain", "example.com")
			Expect(err).To(Succeed())

			icfg := cfg.DefaultInterfaceSettings

			Expect(icfg.Community).To(BeEquivalentTo(crypto.GenerateKeyFromPassword("my-community-password")))
			Expect(icfg.ICE.Username).To(Equal("user1"))
			Expect(icfg.ICE.Password).To(Equal("pass1"))
			Expect(cfg.Backends).To(ConsistOf(
				config.BackendURL{URL: url.URL{Scheme: "p2p"}},
				config.BackendURL{URL: url.URL{Scheme: "grpc", Host: "example.com:8080"}},
			))
			Expect(icfg.ICE.URLs).To(ConsistOf(
				config.URL{url.URL{Scheme: "stun", Opaque: "stun.example.com.:3478"}},
				config.URL{url.URL{Scheme: "stuns", Opaque: "stun.example.com.:3478"}},
				config.URL{url.URL{Scheme: "turn", Opaque: "turn.example.com.:3478", RawQuery: "transport=udp"}},
				config.URL{url.URL{Scheme: "turn", Opaque: "turn.example.com.:3478", RawQuery: "transport=tcp"}},
				config.URL{url.URL{Scheme: "turns", Opaque: "turn.example.com.:5349", RawQuery: "transport=tcp"}},
			))
			Expect(cfg.Interfaces).To(HaveKey("wg-test"))

			err = cfg.Marshal(GinkgoWriter)
			Expect(err).To(Succeed())
		})

		AfterEach(func() {
			dnsSrv.Close()
			webSrv.Close()
			mockdns.UnpatchNet(net.DefaultResolver)
		})
	})
})
