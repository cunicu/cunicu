package socket

import (
	"context"

	"riasc.eu/wice/pkg/pb"
)

func (s *Server) GetStatus(ctx context.Context, _ *pb.Void) (*pb.Status, error) {
	s.daemon.InterfaceLock.Lock()
	defer s.daemon.InterfaceLock.Unlock()

	interfaces := []*pb.Interface{}
	for _, i := range s.daemon.Interfaces {
		interfaces = append(interfaces, i.Marshal())
	}

	return &pb.Status{
		Interfaces: interfaces,
	}, nil
}

func (s *Server) StreamEvents(params *pb.StreamEventsParams, stream pb.Socket_StreamEventsServer) error {
	events := s.daemon.ListenEvents()

	// Send initial connection state of all peers
	for _, i := range s.daemon.Interfaces {
		for key, p := range i.Peers() {
			e := &pb.Event{
				Type:      pb.Event_PEER_CONNECTION_STATE_CHANGED,
				Interface: p.Interface.Name(),
				Peer:      key.Bytes(),
				Event: &pb.Event_PeerConnectionStateChange{
					PeerConnectionStateChange: &pb.PeerConnectionStateChangeEvent{
						NewState: pb.NewConnectionState(p.ConnectionState),
					},
				},
			}

			stream.Send(e)
		}
	}

	for e := range events {
		stream.Send(e)
	}

	return nil
}

func (s *Server) UnWait(ctx context.Context, params *pb.UnWaitParams) (*pb.Error, error) {
	var e = &pb.Error{
		Code:    pb.Error_EALREADY,
		Message: "already unwaited",
	}

	s.waitOnce.Do(func() {
		s.logger.Info("Control socket un-waited")
		s.waitGroup.Done()
		e = pb.Success
	})

	return e, nil
}

func (s *Server) Stop(ctx context.Context, params *pb.StopParams) (*pb.Error, error) {
	if err := s.daemon.Stop(); err != nil {
		return pb.NewError(err), nil
	}

	return pb.Success, nil
}

func (s *Server) Sync(ctx context.Context, params *pb.SyncParams) (*pb.Error, error) {
	if err := s.daemon.SyncAllInterfaces(); err != nil {
		return pb.NewError(err), nil
	}

	return pb.Success, nil
}

func (s *Server) RestartPeer(ctx context.Context, params *pb.RestartPeerParams) (*pb.Error, error) {
	peer, pbErr, err := s.findPeer(params.Intf, params.Peer)
	if pbErr != nil || err != nil {
		return pbErr, err
	}

	peer.Restart()

	return pb.Success, nil
}

func (s *Server) RemoveInterface(ctx context.Context, params *pb.RemoveInterfaceParams) (*pb.Error, error) {
	return pb.NotSupported, nil
}

func (s *Server) SyncInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Error, error) {
	return pb.NotSupported, nil
}

func (s *Server) AddInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Error, error) {
	return pb.NotSupported, nil
}

func (s *Server) SetInterfaceConfig(ctx context.Context, params *pb.InterfaceConfigParams) (*pb.Error, error) {
	return pb.NotSupported, nil
}
