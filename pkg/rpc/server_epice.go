package rpc

import (
	"context"
	"io"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/feat/epdisc"
	icex "github.com/stv0g/cunicu/pkg/feat/epdisc/ice"
	"github.com/stv0g/cunicu/pkg/proto"
	protoepdisc "github.com/stv0g/cunicu/pkg/proto/feat/epdisc"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EndpointDiscoveryServer struct {
	rpcproto.UnimplementedEndpointDiscoverySocketServer

	*Server
	*epdisc.EndpointDiscovery
}

func NewEndpointDiscoveryServer(s *Server, ep *epdisc.EndpointDiscovery) *EndpointDiscoveryServer {
	eps := &EndpointDiscoveryServer{
		Server:            s,
		EndpointDiscovery: ep,
	}

	rpcproto.RegisterEndpointDiscoverySocketServer(s.grpc, eps)

	ep.OnConnectionStateChange(eps)

	return eps
}

func (s *EndpointDiscoveryServer) RestartPeer(ctx context.Context, params *rpcproto.RestartPeerParams) (*proto.Empty, error) {
	pk, err := crypto.ParseKeyBytes(params.Peer)
	if err != nil {
		return &proto.Empty{}, status.Errorf(codes.InvalidArgument, "failed to parse key: %s", err)
	}

	p := s.daemon.Peer(params.Intf, &pk)
	if p == nil {
		return &proto.Empty{}, status.Errorf(codes.NotFound, "unknown peer %s/%s", params.Intf, pk)
	}

	ip := s.Peers[p]
	if ip == nil {
		return &proto.Empty{}, status.Errorf(codes.NotFound, "unknown peer %s/%s", params.Intf, pk)
	}

	err = ip.Restart()
	if err != nil {
		return &proto.Empty{}, status.Errorf(codes.Unknown, "failed to restart peer session: %s", err)
	}

	return &proto.Empty{}, nil
}

func (s *EndpointDiscoveryServer) SendConnectionStates(stream rpcproto.Daemon_StreamEventsServer) {
	for _, p := range s.Peers {
		e := &rpcproto.Event{
			Type:      rpcproto.EventType_PEER_CONNECTION_STATE_CHANGED,
			Interface: p.Interface.Name(),
			Peer:      p.Peer.PublicKey().Bytes(),
			Event: &rpcproto.Event_PeerConnectionStateChange{
				PeerConnectionStateChange: &rpcproto.PeerConnectionStateChangeEvent{
					NewState: protoepdisc.NewConnectionState(p.ConnectionState()),
				},
			},
		}

		if err := stream.Send(e); err == io.EOF {
			continue
		} else if err != nil {
			s.logger.Error("Failed to send", zap.Error(err))
		}
	}
}

func (s *EndpointDiscoveryServer) OnConnectionStateChange(p *epdisc.Peer, new, prev icex.ConnectionState) {
	s.events.Send(&rpcproto.Event{
		Type: rpcproto.EventType_PEER_CONNECTION_STATE_CHANGED,

		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &rpcproto.Event_PeerConnectionStateChange{
			PeerConnectionStateChange: &rpcproto.PeerConnectionStateChangeEvent{
				NewState:  protoepdisc.NewConnectionState(new),
				PrevState: protoepdisc.NewConnectionState(prev),
			},
		},
	})
}

func (s *EndpointDiscoveryServer) InterfaceStatus(ci *core.Interface) *protoepdisc.Interface {
	i, ok := s.Interfaces[ci]
	if !ok {
		return nil
	}

	return i.Marshal()
}

func (s *EndpointDiscoveryServer) PeerStatus(cp *core.Peer) *protoepdisc.Peer {
	p, ok := s.Peers[cp]
	if !ok {
		s.logger.Error("Failed to find peer for", zap.Any("cp", cp.PublicKey()))
		return nil
	}

	return p.Marshal()
}
