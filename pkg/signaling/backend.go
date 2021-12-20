package signaling

import (
	"fmt"
	"io"
	"net/url"

	"riasc.eu/wice/pkg/crypto"
)

var (
	Backends = map[BackendType]*BackendPlugin{}
)

type Backend interface {
	io.Closer

	PublishOffer(kp crypto.PublicKeyPair, offer Offer) error
	SubscribeOffer(kp crypto.PublicKeyPair) (chan Offer, error)
	Tick()
}

func NewBackend(uri *url.URL, options map[string]string) (Backend, error) {
	typ := BackendType(uri.Scheme)

	p, ok := Backends[typ]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %s", typ)
	}

	be, err := p.New(uri, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	return be, nil
}
