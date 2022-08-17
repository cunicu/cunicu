package rpc

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/feat/disc/epice"
	icex "riasc.eu/wice/pkg/ice"
	"riasc.eu/wice/pkg/pb"
)

type EndpointDiscoveryServer struct {
	pb.UnimplementedEndpointDiscoverySocketServer

	*Server
	*epice.EndpointDiscovery
}

func NewEndpointDiscoveryServer(s *Server, ep *epice.EndpointDiscovery) *EndpointDiscoveryServer {
	eps := &EndpointDiscoveryServer{
		Server:            s,
		EndpointDiscovery: ep,
	}

	pb.RegisterEndpointDiscoverySocketServer(s.grpc, eps)

	ep.OnConnectionStateChange(eps)

	return eps
}

func (s *EndpointDiscoveryServer) RestartPeer(ctx context.Context, params *pb.RestartPeerParams) (*pb.Error, error) {
	pk, _ := crypto.ParseKeyBytes(params.Peer)
	p := s.watcher.Peer(params.Intf, &pk)
	if p == nil {
		err := fmt.Errorf("unknown peer %s/%s", params.Intf, pk.String())
		return pb.NewError(err), nil
	}

	ip := s.Peers[p]

	ip.Restart()

	return pb.Success, nil
}

func (s *EndpointDiscoveryServer) SendConnectionStates(stream pb.Socket_StreamEventsServer) {
	for _, p := range s.Peers {
		e := &pb.Event{
			Type:      pb.Event_PEER_CONNECTION_STATE_CHANGED,
			Interface: p.Interface.Name(),
			Peer:      p.Peer.PublicKey().Bytes(),
			Event: &pb.Event_PeerConnectionStateChange{
				PeerConnectionStateChange: &pb.PeerConnectionStateChangeEvent{
					NewState: pb.NewConnectionState(p.ConnectionState),
				},
			},
		}

		if err := stream.Send(e); err != nil {
			s.logger.Error("Failed to send", zap.Error(err))
		}
	}
}

func (s *EndpointDiscoveryServer) OnConnectionStateChange(p *epice.Peer, new, prev icex.ConnectionState) {
	s.events.C <- &pb.Event{
		Type: pb.Event_PEER_CONNECTION_STATE_CHANGED,

		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &pb.Event_PeerConnectionStateChange{
			PeerConnectionStateChange: &pb.PeerConnectionStateChangeEvent{
				NewState:  pb.NewConnectionState(new),
				PrevState: pb.NewConnectionState(prev),
			},
		},
	}
}
