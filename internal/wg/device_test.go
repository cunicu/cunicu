package wg_test

import (
	"bytes"
	"strings"
	"testing"

	"riasc.eu/wice/internal/wg"
)

var config = `[Interface]
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

func TestDeviceConfig(t *testing.T) {
	rd := strings.NewReader(config)

	cfg, err := wg.ParseConfig(rd, "")
	if err != nil {
		t.Fatalf("failed to parse config: %s", err)
	}

	wr := &bytes.Buffer{}
	if err := wg.DumpConfig(wr, cfg); err != nil {
		t.Fatalf("failed to dump config: %s", err)
	}

	if wr.String() != config {
		t.Errorf("configs not equal:\n%s\n%s", config, wr.String())
	}
}
