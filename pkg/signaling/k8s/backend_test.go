package k8s_test

import (
	"log"
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling/k8s"
)

func TestBackend(t *testing.T) {
	uri, err := url.Parse("k8s:?node-name=red")
	if err != nil {
		t.Errorf("failed to parse backend URL: %s", err)
	}

	b, err := k8s.NewBackend(uri, nil)
	if err != nil {
		t.Errorf("failed to create backend: %s", err)
	}

	ourSecretKey, _ := crypto.GeneratePrivateKey()
	theirSecretKey, _ := crypto.GeneratePrivateKey()

	kp := crypto.PublicKeyPair{
		Ours:   ourSecretKey.PublicKey(),
		Theirs: theirSecretKey.PublicKey(),
	}

	o := pb.NewOffer()

	ch, err := b.SubscribeOffer(kp)
	if err != nil {
		t.Errorf("failed to subscribe to offer")
	}

	b.PublishOffer(kp, o)

	n := <-ch
	log.Print(n)
}
