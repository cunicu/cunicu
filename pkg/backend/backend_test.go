package backend_test

import (
	"net/url"
	"testing"

	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/backend/http"
	_ "riasc.eu/wice/pkg/backend/http"
)

func TestNewBackend(t *testing.T) {
	uri, err := url.Parse("http://example.com")
	if err != nil {
		t.Fail()
	}

	b, err := backend.NewBackend(uri, map[string]string{})
	if err != nil {
		t.Fail()
	}

	if _, ok := b.(*http.Backend); !ok {
		t.Fail()
	}
}
