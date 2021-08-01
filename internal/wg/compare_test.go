package wg_test

import (
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/wg"
)

func TestCmpPeersEqual(t *testing.T) {
	a := wgtypes.Peer{}
	b := wgtypes.Peer{}

	if wg.CmpPeers(&a, &a) != 0 {
		t.Fail()
	}

	var err error
	b.PublicKey, err = wgtypes.GenerateKey()
	if err != nil {
		t.Fail()
	}

	if wg.CmpPeers(&a, &b) >= 0 {
		t.Fail()
	}
}
