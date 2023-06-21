// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg_test

import (
	"bytes"
	"net"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/tty"
	"github.com/stv0g/cunicu/pkg/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("device", func() {
	var err error
	var dev wg.Interface
	var sk, skp1, skp2, psk wgtypes.Key

	BeforeEach(func() {
		now := time.Now()

		sk, err = wgtypes.ParseKey("QI1WbUUJJS69sLS1TSBfx5U/n1jMQaPwbcDnq2S24Fg=")
		Expect(err).To(Succeed())

		skp1, err = wgtypes.ParseKey("KOMLkp/ZuAjnrwV2OsXNZ7rx3cBrCjOTv1Zhk1SiDlQ=")
		Expect(err).To(Succeed())

		skp2, err = wgtypes.ParseKey("UL6T4C540jv1xy4cC8nr03wnepLQPDsObRCBhhSzXUM=")
		Expect(err).To(Succeed())

		psk, err = wgtypes.ParseKey("eZuJ5S7fcYVm5wuRZitib4UsqVpmS81hZiZPt5Ob9SE=")
		Expect(err).To(Succeed())

		dev = wg.Interface{
			Name:         "wg0",
			PrivateKey:   sk,
			PublicKey:    sk.PublicKey(),
			ListenPort:   1234,
			FirewallMark: 5678,
			Peers: []wgtypes.Peer{
				{
					PublicKey:    skp1.PublicKey(),
					PresharedKey: psk,
					Endpoint: &net.UDPAddr{
						IP:   net.IPv4(1, 2, 3, 4),
						Port: 51820,
					},
					PersistentKeepaliveInterval: 25 * time.Second,
					AllowedIPs: []net.IPNet{
						{
							IP:   net.IPv4(5, 6, 7, 8),
							Mask: net.CIDRMask(16, 32),
						},
					},
					LastHandshakeTime: now.Add(-5 * time.Second),
				},
				{
					PublicKey:         skp2.PublicKey(),
					LastHandshakeTime: now,
					TransmitBytes:     512,
					ReceiveBytes:      1024,
				},
			},
		}
	})

	It("to config", func() {
		cfg := dev.Config()

		Expect(cfg.PrivateKey).NotTo(BeNil())
		Expect(*cfg.PrivateKey).To(Equal(sk))

		Expect(cfg.ListenPort).NotTo(BeNil())
		Expect(*cfg.ListenPort).To(Equal(1234))

		Expect(cfg.FirewallMark).NotTo(BeNil())
		Expect(*cfg.FirewallMark).To(Equal(5678))

		Expect(cfg.Peers).To(HaveLen(2))
		Expect(cfg.Peers[0].PublicKey).To(Equal(skp1.PublicKey()))

		Expect(cfg.Peers[0].PresharedKey).NotTo(BeNil())
		Expect(*cfg.Peers[0].PresharedKey).To(Equal(psk))

		Expect(cfg.Peers[0].PersistentKeepaliveInterval).NotTo(BeNil())
		Expect(*cfg.Peers[0].PersistentKeepaliveInterval).To(Equal(25 * time.Second))

		Expect(cfg.Peers[0].Endpoint).NotTo(BeNil())
		Expect(cfg.Peers[0].Endpoint.String()).To(Equal("1.2.3.4:51820"))

		Expect(cfg.Peers[0].AllowedIPs).To(HaveLen(1))
		Expect(cfg.Peers[0].AllowedIPs[0].String()).To(Equal("5.6.7.8/16"))
	})

	Context("dump", func() {
		It("hide keys", func() {
			buf := &bytes.Buffer{}
			buf2 := tty.NewANSIStripper(buf)

			err = dev.Dump(buf2, true)
			Expect(err).To(Succeed())

			Expect(buf.String()).Should(Equal(`interface: wg0
  public key: OUE5VJPyG9HEygYZowUJBARyCRIy8joQQKyl/YHvYWc=
  private key: (hidden)
  listening port: 1234
  fwmark: 5678

peer: 6Oh0ZnWPQCVftiiD5P+pLf0c271rBdcQluxYgAGsgj0=
  latest handshake: Now
  allowed ips: (none)
  transfer: 1.00 KiB received, 512 B sent

peer: Y658qGkT02yrLopsu1pnT2/DdgeJdMK8HxDI2UYSOX4=
  preshared key: (hidden)
  endpoint: 1.2.3.4:51820
  latest handshake: 5 seconds ago
  allowed ips: 5.6.7.8/16
  persistent keepalive: every 25 seconds
`))
		})

		It("show keys", func() {
			buf := &bytes.Buffer{}
			buf2 := tty.NewANSIStripper(buf)

			err = dev.Dump(buf2, false)
			Expect(err).To(Succeed())

			Expect(buf.String()).Should(Equal(`interface: wg0
  public key: OUE5VJPyG9HEygYZowUJBARyCRIy8joQQKyl/YHvYWc=
  private key: QI1WbUUJJS69sLS1TSBfx5U/n1jMQaPwbcDnq2S24Fg=
  listening port: 1234
  fwmark: 5678

peer: 6Oh0ZnWPQCVftiiD5P+pLf0c271rBdcQluxYgAGsgj0=
  latest handshake: Now
  allowed ips: (none)
  transfer: 1.00 KiB received, 512 B sent

peer: Y658qGkT02yrLopsu1pnT2/DdgeJdMK8HxDI2UYSOX4=
  preshared key: eZuJ5S7fcYVm5wuRZitib4UsqVpmS81hZiZPt5Ob9SE=
  endpoint: 1.2.3.4:51820
  latest handshake: 5 seconds ago
  allowed ips: 5.6.7.8/16
  persistent keepalive: every 25 seconds
`))
		})
	})
})
