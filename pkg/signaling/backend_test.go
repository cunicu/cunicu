package signaling_test

import (
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/p2p"
)

func TestNewBackend(t *testing.T) {
	uri, err := url.Parse("p2p:")
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}

	events := make(chan *pb.Event, 100)

	cfg := &signaling.BackendConfig{
		URI: uri,
	}

	b, err := signaling.NewBackend(cfg, events)
	if err != nil {
		t.Fatalf("Failed to create new backend: %s", err)
	}

	if _, ok := b.(*p2p.Backend); !ok {
		t.Fail()
	}
}
