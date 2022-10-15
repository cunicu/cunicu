// Package pske uses the Kyber Key Establishment Mechanism (KEM) to establish Preshared Keys (PSKs) between two WireGuard peers
package pske

import (
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/daemon"
	"go.uber.org/zap"
)

func init() {
	daemon.RegisterFeature("pske", "Preshared key establishment ", New, 110)
}

type Interface struct {
	*daemon.Interface

	Peers map[*core.Peer]*Peer

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.EstablishPresharedKeys {
		return nil, nil
	}

	p := &Interface{
		Interface: i,
		Peers:     map[*core.Peer]*Peer{},

		logger: zap.L().Named("pske").With(zap.String("intf", i.Name())),
	}

	i.OnPeer(p)

	return p, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started post-quantum preshared key establishment")

	return nil
}
