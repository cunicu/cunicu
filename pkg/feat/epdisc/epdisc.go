// Package epdisc implements endpoint (EP) discovery using Interactive Connection Establishment (ICE).
package epdisc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	icex "riasc.eu/wice/pkg/feat/epdisc/ice"
	"riasc.eu/wice/pkg/feat/epdisc/proxy"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/watcher"
	"riasc.eu/wice/pkg/wg"
)

type EndpointDiscovery struct {
	Peers      map[*core.Peer]*Peer
	Interfaces map[*core.Interface]*Interface

	onConnectionStateChange []OnConnectionStateHandler

	watcher *watcher.Watcher
	config  *config.Config
	client  *wgctrl.Client
	backend signaling.Backend

	logger *zap.Logger
}

func New(w *watcher.Watcher, cfg *config.Config, client *wgctrl.Client, backend signaling.Backend) *EndpointDiscovery {
	e := &EndpointDiscovery{
		Peers:      map[*core.Peer]*Peer{},
		Interfaces: map[*core.Interface]*Interface{},

		onConnectionStateChange: []OnConnectionStateHandler{},

		watcher: w,
		config:  cfg,
		client:  client,
		backend: backend,

		logger: zap.L().Named("epdisc"),
	}

	w.OnAll(e)

	return e
}

func (e *EndpointDiscovery) Start() error {
	e.logger.Info("Started endpoint discovery")

	return nil
}

func (e *EndpointDiscovery) Close() error {
	// First switch all sessions to closing so they dont get restarted
	for _, p := range e.Peers {
		p.setConnectionState(icex.ConnectionStateClosing)
	}

	for _, p := range e.Peers {
		if err := p.Close(); err != nil {
			return fmt.Errorf("failed to close peer: %w", err)
		}
	}

	for _, i := range e.Interfaces {
		if err := i.Close(); err != nil {
			return fmt.Errorf("failed to close interface: %w", err)
		}
	}

	return nil
}

func (e *EndpointDiscovery) OnInterfaceAdded(ci *core.Interface) {
	i, err := NewInterface(ci, e)
	if err != nil {
		e.logger.Error("Failed to initialize ICE interface", zap.Error(err))
		return
	}

	e.Interfaces[ci] = i
}

func (e *EndpointDiscovery) OnInterfaceRemoved(ci *core.Interface) {
	i := e.Interfaces[ci]

	if err := i.Close(); err != nil {
		e.logger.Error("Failed to de-initialize ICE interface", zap.Error(err))
	}

	delete(e.Interfaces, ci)
}

func (e *EndpointDiscovery) OnInterfaceModified(ci *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	i := e.Interfaces[ci]

	if m.Is(core.InterfaceModifiedListenPort) {
		if err := i.UpdateRedirects(); err != nil {
			e.logger.Error("Failed to update DPAT redirects", zap.Error(err))
		}
	}

	for _, cp := range i.Peers {
		p := e.Peers[cp]

		if m.Is(core.InterfaceModifiedListenPort) {
			if kproxy, ok := p.proxy.(*proxy.KernelProxy); ok {
				if err := kproxy.UpdateListenPort(i.ListenPort); err != nil {
					e.logger.Error("Failed to update SPAT redirect", zap.Error(err))
				}
			}
		}

		if m.Is(core.InterfaceModifiedPrivateKey) {

			skOld := crypto.Key(old.PrivateKey)
			if err := p.Resubscribe(context.Background(), skOld); err != nil {
				e.logger.Error("Failed to update subscription", zap.Error(err))
			}
		}
	}
}

func (e *EndpointDiscovery) OnPeerAdded(cp *core.Peer) {
	i := e.Interfaces[cp.Interface]

	p, err := NewPeer(cp, i)
	if err != nil {
		e.logger.Error("Failed to initialize ICE peer", zap.Error(err))
		return
	}

	e.Peers[cp] = p
}

func (e *EndpointDiscovery) OnPeerRemoved(cp *core.Peer) {
	p := e.Peers[cp]

	if err := p.Close(); err != nil {
		e.logger.Error("Failed to de-initialize ICE peer", zap.Error(err))
	}

	delete(e.Peers, cp)
}

func (e *EndpointDiscovery) OnPeerModified(cp *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	p := e.Peers[cp]

	// TODO: Handle changed endpoint addresses
	//       What do we do when they have been changed externally?
	if m.Is(core.PeerModifiedEndpoint) {
		// Check if change was external
		epNew := p.Endpoint
		epExpected := p.lastEndpoint

		if (epExpected != nil && epNew != nil) && (!epNew.IP.Equal(epExpected.IP) || epNew.Port != epExpected.Port) {
			e.logger.Warn("Endpoint address has been changed externally. This is breaks the connection and is most likely not desired.")
		}
	}
}
