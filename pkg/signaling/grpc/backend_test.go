package grpc_test

import (
	"net"
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling/grpc"
)

func newOffer() *pb.Offer {
	return &pb.Offer{
		Version: 456,
		Epoch:   789,
		Type:    pb.Offer_OFFER,
		Role:    pb.Offer_CONTROLLING,
	}
}

func TestBackend(t *testing.T) {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatalf("Failed to listen: %s", err)
	}

	svr := grpc.NewServer()

	go svr.Serve(l)
	defer svr.Stop()

	uri, err := url.Parse("grpc://127.0.0.1:8080?insecure=true")
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}

	be, err := grpc.NewBackend(uri, nil)
	if err != nil {
		t.Fatalf("Failed to create backend: %s", err)
	}
	defer be.Close()

	so := newOffer()

	ourKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.FailNow()
	}

	theirKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		t.FailNow()
	}

	kp := crypto.KeyPair{
		Ours:   ourKey.PublicKey(),
		Theirs: theirKey.PublicKey(),
	}

	ch, err := be.SubscribeOffers(kp)
	if err != nil {
		t.Fatalf("Failed to subscribe to offers: %s", err)
	}

	if err := be.PublishOffer(kp, so); err != nil {
		t.Fatalf("Failed to publish offer: %s", err)
	}

	ro := <-ch

	if so.Epoch != ro.Epoch || so.Version != ro.Version {
		t.Fatalf("Offer mismatch: %v != %v", so, ro)
	}
}
