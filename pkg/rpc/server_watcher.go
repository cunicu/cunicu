package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *WatcherServer) GetStatus(ctx context.Context, _ *pb.Empty) (*pb.Status, error) {
	s.InterfaceLock.Lock()
	defer s.InterfaceLock.Unlock()

	interfaces := []*pb.Interface{}
	for _, i := range s.Interfaces {
		interfaces = append(interfaces, i.Marshal())
	}

	return &pb.Status{
		Interfaces: interfaces,
	}, nil
}

func (s *WatcherServer) Sync(ctx context.Context, params *pb.SyncParams) (*pb.Empty, error) {
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
