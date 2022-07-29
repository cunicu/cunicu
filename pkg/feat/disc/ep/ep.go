package ep

import (
	"net"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/core"
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

func New(w *watcher.Watcher, cfg *config.Config, client *wgctrl.Client, backend signaling.Backend) (*EndpointDiscovery, error) {
	e := &EndpointDiscovery{
		Peers:      map[*core.Peer]*Peer{},
		Interfaces: map[*core.Interface]*Interface{},

		onConnectionStateChange: []OnConnectionStateHandler{},

		watcher: w,
		config:  cfg,
		client:  client,
		backend: backend,

		logger: zap.L().Named("ep-disc"),
	}

	w.OnAll(e)

	return e, nil
}

func (e *EndpointDiscovery) OnConnectionStateChange(h OnConnectionStateHandler) {
	e.onConnectionStateChange = append(e.onConnectionStateChange, h)
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
	// i := e.interfaces[ci]

	// TODO: Handle changed listen port
	// if m&core.InterfaceModifiedListenPort != 0 {}
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
	// p := e.peers[cp]

	// TODO: Handle changed endpoint addresses
	//       What do we do when they have been changed externally?
	// if m&core.PeerModifiedEndpoint != 0 {}
}
