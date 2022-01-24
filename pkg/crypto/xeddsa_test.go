package crypto_test

import (
	"crypto/rand"
	"fmt"
	"testing"

	"riasc.eu/wice/pkg/crypto"
)

func TestXEdDSA(t *testing.T) {
	sk, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fail()
	}

	pk := sk.PublicKey()

	msg := make([]byte, 200)

	nonce, err := crypto.GetNonce()
	if err != nil {
		t.Fail()
	}

	_, err = rand.Reader.Read(msg[:])
	if err != nil {
		t.Fail()
	}

	signature := sk.Sign(msg, nonce)

	fmt.Printf("PrivateKey = %s\n", sk)
	fmt.Printf("PublicKey = %s\n", pk)
	fmt.Printf("Signature = %s\n", signature)

	res := pk.Verify(msg, signature)
	if !res {
		t.Error("Signature mismatch")
	}

	msg[0] ^= 0xff

	res = pk.Verify(msg, signature)
	if res {
		t.Error("Signature false positive")
	}
}
