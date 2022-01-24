package signaling

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

var (
	Backends = map[BackendType]*BackendPlugin{}
)

type BackendType string // URL schemes

type BackendFactory func(*url.URL, chan *pb.Event) (Backend, error)

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

func NewBackend(uri *url.URL, events chan *pb.Event) (Backend, error) {
	typs := strings.SplitN(uri.Scheme, "+", 2)
	typ := BackendType(typs[0])

	p, ok := Backends[typ]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %s", typ)
	}

	if len(typs) > 1 {
		uri.Scheme = typs[1]
	}

	be, err := p.New(uri, events)
	if err != nil {
		return nil, err
	}

	return be, nil
}
