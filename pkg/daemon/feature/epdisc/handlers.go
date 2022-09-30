package epdisc

import (
	"context"
	"net"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc/proxy"
	"github.com/stv0g/cunicu/pkg/wg"

	icex "github.com/stv0g/cunicu/pkg/ice"
)

type OnConnectionStateHandler interface {
	OnConnectionStateChange(p *Peer, new, prev icex.ConnectionState)
}

func (e *Interface) OnConnectionStateChange(h OnConnectionStateHandler) {
	e.onConnectionStateChange = append(e.onConnectionStateChange, h)
}

func (e *Interface) OnInterfaceModified(ci *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	if m.Is(core.InterfaceModifiedListenPort) {
		if err := e.UpdateRedirects(); err != nil {
			e.logger.Error("Failed to update DPAT redirects", zap.Error(err))
		}
	}

	for _, p := range e.Peers {
		if m.Is(core.InterfaceModifiedListenPort) {
			if kproxy, ok := p.proxy.(*proxy.KernelProxy); ok {
				if err := kproxy.UpdateListenPort(ci.ListenPort); err != nil {
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

func (e *Interface) OnPeerAdded(cp *core.Peer) {
	p, err := NewPeer(cp, e)
	if err != nil {
		e.logger.Error("Failed to initialize ICE peer", zap.Error(err))
		return
	}

	e.Peers[cp] = p
}

func (e *Interface) OnPeerRemoved(cp *core.Peer) {
	p, ok := e.Peers[cp]
	if !ok {
		return
	}

	if err := p.Close(); err != nil {
		e.logger.Error("Failed to de-initialize ICE peer", zap.Error(err))
	}

	delete(e.Peers, cp)
}

func (e *Interface) OnPeerModified(cp *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
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
