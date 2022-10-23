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

func (i *Interface) OnConnectionStateChange(h OnConnectionStateHandler) {
	i.onConnectionStateChange = append(i.onConnectionStateChange, h)
}

func (i *Interface) OnInterfaceModified(ci *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	if m.Is(core.InterfaceModifiedListenPort) {
		if err := i.UpdateRedirects(); err != nil {
			i.logger.Error("Failed to update DPAT redirects", zap.Error(err))
		}
	}

	for _, p := range i.Peers {
		if m.Is(core.InterfaceModifiedListenPort) {
			if kproxy, ok := p.proxy.(*proxy.KernelProxy); ok {
				if err := kproxy.UpdateListenPort(ci.ListenPort); err != nil {
					i.logger.Error("Failed to update SPAT redirect", zap.Error(err))
				}
			}
		}

		if m.Is(core.InterfaceModifiedPrivateKey) {
			skOld := crypto.Key(old.PrivateKey)
			if err := p.Resubscribe(context.Background(), skOld); err != nil {
				i.logger.Error("Failed to update subscription", zap.Error(err))
			}
		}
	}
}

func (i *Interface) OnPeerAdded(cp *core.Peer) {
	p, err := NewPeer(cp, i)
	if err != nil {
		i.logger.Error("Failed to initialize ICE peer", zap.Error(err))
		return
	}

	i.Peers[cp] = p
}

func (i *Interface) OnPeerRemoved(cp *core.Peer) {
	p, ok := i.Peers[cp]
	if !ok {
		return
	}

	if err := p.Close(); err != nil {
		i.logger.Error("Failed to de-initialize ICE peer", zap.Error(err))
	}

	delete(i.Peers, cp)
}

func (i *Interface) OnPeerModified(cp *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	p := i.Peers[cp]

	if m.Is(core.PeerModifiedEndpoint) {
		// Check if change was external
		epNew := p.Endpoint
		epExpected := p.lastEndpoint

		if (epExpected != nil && epNew != nil) && (!epNew.IP.Equal(epExpected.IP) || epNew.Port != epExpected.Port) {
			i.logger.Warn("Endpoint address has been changed externally. This is breaks the connection and is most likely not desired.")
		}
	}
}
