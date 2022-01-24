package pb_test

import (
	"testing"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

func TestSignedMessage(t *testing.T) {
	o := pb.Offer{
		Version: 1234,
	}

	sk, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %s", err)
	}

	pk := sk.PublicKey()

	sm := pb.SignedMessage{}
	if err := sm.Marshal(&o, sk); err != nil {
		t.Fatalf("Failed to sign message: %s", err)
	}

	o2 := pb.Offer{}
	if err := sm.Unmarshal(&o2, pk); err != nil {
		t.Fatalf("Failed to validate message: %s", err)
	}

	if o.Version != o2.Version {
		t.Fatal("Mismatch")
	}

	sm.Body[0] ^= 1

	if err := sm.Unmarshal(&o2, pk); err == nil {
		t.Fatal("Validated invalid message")
	}
}

func TestEncryptedMessage(t *testing.T) {
	o := pb.Offer{
		Version: 1234,
	}

	sk, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("failed to generate key: %s", err)
	}

	pk := sk.PublicKey()

	em := pb.EncryptedMessage{}
	if err := em.Marshal(&o, pk); err != nil {
		t.Fatalf("Failed to encrypto message: %s", err)
	}

	o2 := pb.Offer{}
	if err := em.Unmarshal(&o2, sk); err != nil {
		t.Fatalf("Failed to decrypt message: %s", err)
	}

	em.Body[0] ^= 1

	if err := em.Unmarshal(&o2, sk); err == nil {
		t.Fatalf("Decrypted invalid message: %s", err)
	}
}
