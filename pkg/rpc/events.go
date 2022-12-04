package rpc

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/wg"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (s *Server) OnInterfaceAdded(i *core.Interface) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_INTERFACE_ADDED,
		Interface: i.Name(),
	})
}

func (s *Server) OnInterfaceRemoved(i *core.Interface) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_INTERFACE_REMOVED,
		Interface: i.Name(),
	})
}

func (s *Server) OnInterfaceModified(i *core.Interface, old *wg.Device, mod core.InterfaceModifier) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_INTERFACE_MODIFIED,
		Interface: i.Name(),
		Event: &rpcproto.Event_InterfaceModified{
			InterfaceModified: &rpcproto.InterfaceModifiedEvent{
				Modified: uint32(mod),
			},
		},
	})
}

func (s *Server) OnPeerAdded(p *core.Peer) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_PEER_ADDED,
		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),
	})
}

func (s *Server) OnPeerRemoved(p *core.Peer) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_PEER_REMOVED,
		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),
	})
}

func (s *Server) OnPeerModified(p *core.Peer, old *wgtypes.Peer, mod core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_PEER_MODIFIED,
		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &rpcproto.Event_PeerModified{
			PeerModified: &rpcproto.PeerModifiedEvent{
				Modified: uint32(mod),
			},
		},
	})
}

func (s *Server) OnSignalingBackendReady(b signaling.Backend) {
	s.events.Send(&rpcproto.Event{
		Type: rpcproto.EventType_BACKEND_READY,

		Event: &rpcproto.Event_BackendReady{
			BackendReady: &rpcproto.SignalingBackendReadyEvent{
				Type: b.Type(),
			},
		},
	})
}

func (s *Server) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
}
