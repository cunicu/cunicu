package signaling_test

import (
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/p2p"
)

func TestNewBackend(t *testing.T) {
	uri, err := url.Parse("http://example.com")
	if err != nil {
		t.Fail()
	}

	b, err := signaling.NewBackend(uri, map[string]string{})
	if err != nil {
		t.Fail()
	}

	if _, ok := b.(*p2p.Backend); !ok {
		t.Fail()
	}
}
