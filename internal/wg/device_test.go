package wg_test

import (
	"bytes"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/wg"
)

var _ = Context("device", func() {
	const config = `[Interface]
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

[Peer]
PublicKey           = mBgUyqcI0XXrWskB5w9Z+C3LX5Gu5kw4mDTFPigu/Xg=
AllowedIPs          = 0.0.0.0/0, ::/0
Endpoint            = 14.10.19.13:3436
PersistentKeepalive = 25
`

	Specify("check config parsing and serialization", func() {
		rd := strings.NewReader(config)

		cfg, err := wg.ParseConfig(rd, "")
		Expect(err).To(Succeed(), "failed to parse config: %s", err)

		wr := &bytes.Buffer{}
		err = cfg.Dump(wr)
		Expect(err).To(Succeed(), "failed to dump config: %s", err)

		Expect(wr.String()).To(Equal(config), "configs not equal:\n%s\n%s", config, wr.String())
	})
})
