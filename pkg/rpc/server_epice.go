package rpc

import (
	"context"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *EndpointDiscoveryServer) RestartPeer(ctx context.Context, params *pb.RestartPeerParams) (*pb.Empty, error) {
	pk, err := crypto.ParseKeyBytes(params.Peer)
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.InvalidArgument, "failed to parse key: %s", err)
	}

	p := s.watcher.Peer(params.Intf, &pk)
	if p == nil {
		return &pb.Empty{}, status.Errorf(codes.NotFound, "unknown peer %s/%s", params.Intf, pk.String())
	}

	ip := s.Peers[p]
	if ip == nil {
		return &pb.Empty{}, status.Errorf(codes.NotFound, "unknown peer %s/%s", params.Intf, pk.String())
	}

	err = ip.Restart()
	if err != nil {
		return &pb.Empty{}, status.Errorf(codes.Unknown, "failed to restart peer session: %s", err)
	}

	return &pb.Empty{}, nil
}

func (s *EndpointDiscoveryServer) SendConnectionStates(stream pb.Socket_StreamEventsServer) {
	for _, p := range s.Peers {
		e := &pb.Event{
			Type:      pb.Event_PEER_CONNECTION_STATE_CHANGED,
			Interface: p.Interface.Name(),
			Peer:      p.Peer.PublicKey().Bytes(),
			Event: &pb.Event_PeerConnectionStateChange{
				PeerConnectionStateChange: &pb.PeerConnectionStateChangeEvent{
					NewState: pb.NewConnectionState(p.ConnectionState()),
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

func (s *EndpointDiscoveryServer) OnConnectionStateChange(p *epice.Peer, new, prev icex.ConnectionState) {
	s.events.Send(&pb.Event{
		Type: pb.Event_PEER_CONNECTION_STATE_CHANGED,

		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &pb.Event_PeerConnectionStateChange{
			PeerConnectionStateChange: &pb.PeerConnectionStateChangeEvent{
				NewState:  pb.NewConnectionState(new),
				PrevState: pb.NewConnectionState(prev),
			},
		},
	})
}
