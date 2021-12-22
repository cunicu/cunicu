package signaling_test

import (
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/p2p"
	"riasc.eu/wice/pkg/socket"
)

func TestNewBackend(t *testing.T) {
	uri, err := url.Parse("p2p:")
	if err != nil {
		t.Fatalf("Failed to parse URL: %s", err)
	}

	s, err := socket.Listen("tcp4", "127.0.0.1:0", false)
	if err != nil {
		t.Fatalf("Failed to listen for control socket: %s", err)
	}

	b, err := signaling.NewBackend(uri, s)
	if err != nil {
		t.Fatalf("Failed to create new backend: %s", err)
	}

	if _, ok := b.(*p2p.Backend); !ok {
		t.Fail()
	}
}
