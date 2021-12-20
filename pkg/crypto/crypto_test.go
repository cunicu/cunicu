package crypto_test

import (
	"bytes"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/pkg/crypto"
)

func TestCurve25519Crypt(t *testing.T) {
	keyA, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	keyB, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	pubA := keyA.PublicKey()
	pubB := keyB.PublicKey()

	payload, err := util.GenerateRandomBytes(32)
	if err != nil {
		t.Fail()
	}

	encPayload, err := crypto.Curve25519Crypt(crypto.Key(keyA), crypto.Key(pubB), payload)
	if err != nil {
		t.Fail()
	}

	decPayload, err := crypto.Curve25519Crypt(crypto.Key(keyB), crypto.Key(pubA), encPayload)
	if err != nil {
		t.Fail()
	}

	if !bytes.Equal(decPayload, payload) {
		t.Fail()
	}
}
