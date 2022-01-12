package k8s_test

import (
	"log"
	"net/url"
	"os"
	"testing"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling/k8s"
)

func TestBackend(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skipf("Kubernetes tests are not yet supported in CI")
	}

	uri, err := url.Parse("k8s:?node=red")
	if err != nil {
		t.Errorf("failed to parse backend URL: %s", err)
	}

	events := make(chan *pb.Event, 100)

	b, err := k8s.NewBackend(uri, events)
	if err != nil {
		t.Errorf("failed to create backend: %s", err)
	}

	ourSecretKey, _ := crypto.GeneratePrivateKey()
	theirSecretKey, _ := crypto.GeneratePrivateKey()

	kp := crypto.PublicKeyPair{
		Ours:   ourSecretKey.PublicKey(),
		Theirs: theirSecretKey.PublicKey(),
	}

	o := &pb.Offer{}

	ch, err := b.SubscribeOffer(kp)
	if err != nil {
		t.Errorf("failed to subscribe to offer")
	}

	b.PublishOffer(kp, o)

	n := <-ch
	log.Print(n)
}
