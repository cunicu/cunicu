package config_test

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/foxcpp/go-mockdns"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pion/ice/v2"
	"riasc.eu/wice/pkg/config"

	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
)

var _ = Describe("lookup", func() {
	var dnsSrv *mockdns.Server
	var webSrv *ghttp.Server

	var cfgPath = "/wice"

	BeforeEach(func() {
		var err error

		webSrv = ghttp.NewServer()
		webSrv.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", cfgPath),
				ghttp.VerifyHeader(http.Header{
					"User-agent": []string{"wice"},
				}),
				ghttp.RespondWith(http.StatusOK, "wireguard: { interfaces: [ wg-test ] }",
					http.Header{
						"Content-type": []string{"text/yaml"},
					}),
			),
		)

		dnsSrv, err = mockdns.NewServerWithLogger(map[string]mockdns.Zone{
			"example.com.": {
				A: []string{"1.2.3.4"},
				TXT: []string{
					"wice-backend=p2p",
					"wice-backend=grpc://example.com:8080",
					"wice-community=my-community-password",
					"wice-ice-username=user1",
					"wice-ice-password=pass1",
					fmt.Sprintf("wice-config=%s%s", webSrv.URL(), cfgPath),
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

	It("can do DNS auto configuration", Label("dns-auto-config"), func() {
		c, err := config.ParseArgs("--domain", "example.com")

		Expect(err).To(Succeed())
		Expect(c.PeerDisc.Community).To(Equal("my-community-password"))
		Expect(c.EndpointDisc.ICE.Username).To(Equal("user1"))
		Expect(c.EndpointDisc.ICE.Password).To(Equal("pass1"))
		Expect(c.Backends).To(ConsistOf(
			config.BackendURL{URL: url.URL{Scheme: "p2p"}},
			config.BackendURL{URL: url.URL{Scheme: "grpc", Host: "example.com:8080"}},
		))
		Expect(c.EndpointDisc.ICE.URLs).To(ConsistOf(
			icex.URL{URL: ice.URL{Scheme: ice.SchemeTypeSTUN, Host: "stun.example.com.", Port: 3478, Proto: ice.ProtoTypeUDP}},
			icex.URL{URL: ice.URL{Scheme: ice.SchemeTypeTURN, Host: "turn.example.com.", Port: 3478, Proto: ice.ProtoTypeUDP}},
			icex.URL{URL: ice.URL{Scheme: ice.SchemeTypeSTUNS, Host: "stun.example.com.", Port: 3478, Proto: ice.ProtoTypeTCP}},
			icex.URL{URL: ice.URL{Scheme: ice.SchemeTypeTURNS, Host: "turn.example.com.", Port: 5349, Proto: ice.ProtoTypeTCP}},
			icex.URL{URL: ice.URL{Scheme: ice.SchemeTypeTURN, Host: "turn.example.com.", Port: 3478, Proto: ice.ProtoTypeTCP}},
		))
		Expect(c.WireGuard.Interfaces).To(ContainElement("wg-test"))

		c.Dump(GinkgoWriter)
	})

	AfterEach(func() {
		dnsSrv.Close()
		webSrv.Close()
		mockdns.UnpatchNet(net.DefaultResolver)
	})
})
