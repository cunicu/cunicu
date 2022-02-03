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

type BackendFactory func(*BackendConfig, chan *pb.Event) (Backend, error)

type BackendPlugin struct {
	New         BackendFactory
	Description string
}

type BackendConfig struct {
	URI       *url.URL
	Community *crypto.Key
}

type Backend interface {
	io.Closer

	Publish(kp *crypto.KeyPair, msg *pb.SignalingMessage) error
	Subscribe(kp *crypto.KeyPair) (chan *pb.SignalingMessage, error)
}

func NewBackend(cfg *BackendConfig, events chan *pb.Event) (Backend, error) {
	typs := strings.SplitN(cfg.URI.Scheme, "+", 2)
	typ := BackendType(typs[0])

	p, ok := Backends[typ]
	if !ok {
		return nil, fmt.Errorf("unknown backend type: %s", typ)
	}

	if len(typs) > 1 {
		cfg.URI.Scheme = typs[1]
	}

	be, err := p.New(cfg, events)
	if err != nil {
		return nil, err
	}

	return be, nil
}
