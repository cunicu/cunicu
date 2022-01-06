package signaling

import (
	"fmt"
	"io"
	"net/url"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/socket"
)

var (
	Backends = map[BackendType]*BackendPlugin{}
)

type BackendType string // URL schemes

type BackendFactory func(*url.URL, *socket.Server) (Backend, error)

type BackendPlugin struct {
	New         BackendFactory
	Description string
}

type Backend interface {
	io.Closer

	PublishOffer(kp crypto.PublicKeyPair, offer *pb.Offer) error
	SubscribeOffer(kp crypto.PublicKeyPair) (chan *pb.Offer, error)
	Tick()
}

func NewBackend(uri *url.URL, server *socket.Server) (Backend, error) {
	typ := BackendType(uri.Scheme)

	p, ok := Backends[typ]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %s", typ)
	}

	be, err := p.New(uri, server)
	if err != nil {
		return nil, fmt.Errorf("failed to create backend: %w", err)
	}

	return be, nil
}
