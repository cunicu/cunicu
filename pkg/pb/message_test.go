package pb_test

import (
	"testing"

	"riasc.eu/wice/internal/test"
	"riasc.eu/wice/pkg/pb"
)

func TestEncryptedMessage(t *testing.T) {
	sd := pb.SessionDescription{
		Epoch: 1234,
	}

	ourKP, theirKP, err := test.GenerateKeyPairs()
	if err != nil {
		t.FailNow()
	}

	em := pb.EncryptedMessage{}
	if err := em.Marshal(&sd, ourKP); err != nil {
		t.Fatalf("Failed to encrypto message: %s", err)
	}

	sd2 := pb.SessionDescription{}
	if err := em.Unmarshal(&sd2, theirKP); err != nil {
		t.Fatalf("Failed to decrypt message: %s", err)
	}

	em.Body[0] ^= 1

	if err := em.Unmarshal(&sd2, theirKP); err == nil {
		t.Fatalf("Decrypted invalid message: %s", err)
	}
}
