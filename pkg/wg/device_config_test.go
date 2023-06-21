// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg_test

import (
	"bytes"

	"github.com/stv0g/cunicu/pkg/wg"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("device config", func() {
	var err error
	var cfg *wg.Config

	var cfgStr string

	BeforeEach(func() {
		cfgStr = `[Interface]
PrivateKey = 6Hw0A9Cv0LuCzbdwxPsrmW8oPvOyiyVelwH2pqKlAFE=
ListenPort = 51823
FwMark     = 4096
Address    = 192.168.0.1/24, fd00::1/64
DNS        = 1.1.1.1
MTU        = 1420
Table      = off
PreUp      = ip addr add 2a09:11c0:200::5 peer 2a09:11c0:200::4 dev %i
PostUp     = ip addr add 172.23.156.5 peer 172.23.156.4 dev %i
PostUp     = ip addr add fe80::5/64 dev %i
PostDown   = bla1
PostDown   = bla2
`
	})

	JustBeforeEach(func() {
		cfg, err = wg.ParseConfig([]byte(cfgStr))
		Expect(err).To(Succeed(), "failed to parse config: %s", err)
	})

	test := func() {
		It("can parse and serialize", func() {
			wr := &bytes.Buffer{}
			err = cfg.Dump(wr)
			Expect(err).To(Succeed(), "failed to dump config: %s", err)

			Expect(wr.String()).To(Equal(cfgStr), "configs not equal:\n%s\n%s", cfgStr, wr)
		})
	}

	test()

	When("it has a peer", func() {
		BeforeEach(func() {
			cfgStr += `
# de-fra-1
[Peer]
PublicKey           = mBgUyqcI0XXrWskB5w9Z+C3LX5Gu5kw4mDTFPigu/Xg=
PresharedKey        = zrD9FH+NTECIf7gcpiuvrC4qD2sY2a4YN7fjPcI+RQ8=
AllowedIPs          = 0.0.0.0/0, ::/0
Endpoint            = 14.10.19.13:3436
PersistentKeepalive = 25
`
		})

		test()
	})
})
