// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/wg"
)

func (s *Server) OnInterfaceAdded(i *daemon.Interface) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_INTERFACE_ADDED,
		Interface: i.Name(),
	})
}

func (s *Server) OnInterfaceRemoved(i *daemon.Interface) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_INTERFACE_REMOVED,
		Interface: i.Name(),
	})
}

func (s *Server) OnInterfaceModified(i *daemon.Interface, _ *wg.Interface, mod daemon.InterfaceModifier) {
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

func (s *Server) OnPeerAdded(p *daemon.Peer) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_PEER_ADDED,
		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),
	})
}

func (s *Server) OnPeerRemoved(p *daemon.Peer) {
	s.events.Send(&rpcproto.Event{
		Type:      rpcproto.EventType_PEER_REMOVED,
		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),
	})
}

func (s *Server) OnPeerModified(p *daemon.Peer, _ *wgtypes.Peer, mod daemon.PeerModifier, _, _ []net.IPNet) {
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

func (s *Server) OnSignalingMessage(_ *crypto.PublicKeyPair, _ *signaling.Message) {
}
