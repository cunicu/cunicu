package rpc

import (
	"context"

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

func (s *WatcherServer) GetStatus(ctx context.Context, _ *pb.Void) (*pb.Status, error) {
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

func (s *WatcherServer) Sync(ctx context.Context, params *pb.SyncParams) (*pb.Error, error) {
	if err := s.Watcher.Sync(); err != nil {
		return pb.NewError(err), nil
	}

	return pb.Success, nil
}

func (s *WatcherServer) RemoveInterface(ctx context.Context, params *pb.RemoveInterfaceParams) (*pb.Error, error) {
	return pb.ErrNotSupported, nil
}

func (s *WatcherServer) SyncInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Error, error) {
	return pb.ErrNotSupported, nil
}

func (s *WatcherServer) AddInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Error, error) {
	return pb.ErrNotSupported, nil
}

func (s *WatcherServer) SetInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Error, error) {
	return pb.ErrNotSupported, nil
}
