package k8s_test

import (
	"log"
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/k8s"
)

func TestBackend(t *testing.T) {
	opts := map[string]string{
		"nodename": "red",
	}

	uri, err := url.Parse("k8s://")
	if err != nil {
		t.Errorf("failed to parse backend URL: %w", err)
	}

	b, err := k8s.NewBackend(uri, opts)
	if err != nil {
		t.Errorf("failed to create backend: %w", err)
	}

	ourSecretKey, _ := crypto.GeneratePrivateKey()
	theirSecretKey, _ := crypto.GeneratePrivateKey()

	kp := crypto.PublicKeyPair{
		Ours:   ourSecretKey.PublicKey(),
		Theirs: theirSecretKey.PublicKey(),
	}

	o := signaling.NewOffer()

	ch, err := b.SubscribeOffer(kp)
	if err != nil {
		t.Errorf("failed to subscribe to offer")
	}

	b.PublishOffer(kp, o)

	n := <-ch
	log.Print(n)
}
