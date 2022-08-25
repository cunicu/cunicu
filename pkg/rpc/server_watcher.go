package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/watcher"
)

type WatcherServer struct {
	pb.UnimplementedWatcherServer

	*Server
	*watcher.Watcher
}

func NewWatcherServer(s *Server, w *watcher.Watcher) *WatcherServer {
	ws := &WatcherServer{
		Server:  s,
		Watcher: w,
	}

	pb.RegisterWatcherServer(s.grpc, ws)

	w.OnAll(s)

	return ws
}

func (s *WatcherServer) GetStatus(ctx context.Context, p *pb.StatusParams) (*pb.StatusResp, error) {
	var err error
	var pk crypto.Key

	s.InterfaceLock.Lock()
	defer s.InterfaceLock.Unlock()

	if p.Peer != nil {
		if pk, err = crypto.ParseKeyBytes(p.Peer); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid peer key")
		}
	}

	qis := []*pb.Interface{}
	for _, ci := range s.Interfaces {
		if p.Intf != "" && ci.Name() != p.Intf {
			continue
		}

		qi := ci.MarshalWithPeers(func(cp *core.Peer) *pb.Peer {
			if pk.IsSet() && pk != cp.PublicKey() {
				return nil
			}

			qp := cp.Marshal()

			if s.epice != nil {
				qp.Ice = s.epice.PeerStatus(cp)
			}

			return qp
		})

		qis = append(qis, qi)
	}

	// Check if filters matched anything
	if p.Intf != "" && len(qis) == 0 {
		return nil, status.Errorf(codes.NotFound, "no such interface '%s'", p.Intf)
	} else if pk.IsSet() && len(qis[0].Peers) == 0 {
		return nil, status.Errorf(codes.NotFound, "no such peer '%s' for interface '%s'", pk, p.Intf)
	}

	return &pb.StatusResp{
		Interfaces: qis,
	}, nil
}

func (s *WatcherServer) Sync(ctx context.Context, params *pb.Empty) (*pb.Empty, error) {
	if err := s.Watcher.Sync(); err != nil {
		return &pb.Empty{}, status.Errorf(codes.Unknown, "failed to sync: %s", err)
	}

	return &pb.Empty{}, nil
}

func (s *WatcherServer) RemoveInterface(ctx context.Context, params *pb.RemoveInterfaceParams) (*pb.Empty, error) {
	return &pb.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *WatcherServer) SyncInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Empty, error) {
	return &pb.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *WatcherServer) AddInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Empty, error) {
	return &pb.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *WatcherServer) SetInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Empty, error) {
	return &pb.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}
