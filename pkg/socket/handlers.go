package socket

import (
	"context"
	"fmt"

	"riasc.eu/wice/pkg/pb"
)

func (s *Server) GetStatus(ctx context.Context, _ *pb.Void) (*pb.Status, error) {
	s.daemon.Watcher.InterfaceLock.Lock()
	defer s.daemon.Watcher.InterfaceLock.Unlock()

	interfaces := []*pb.Interface{}
	for _, i := range s.daemon.Watcher.Interfaces {
		interfaces = append(interfaces, i.Marshal())
	}

	return &pb.Status{
		Interfaces: interfaces,
	}, nil
}

func (s *Server) sendConnectionStates(stream pb.Socket_StreamEventsServer) {
	for _, p := range s.daemon.EndpointDiscovery.Peers {
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

		stream.Send(e)
	}
}

func (s *Server) StreamEvents(params *pb.StreamEventsParams, stream pb.Socket_StreamEventsServer) error {
	// Send initial connection state of all peers
	s.sendConnectionStates(stream)

	for e := range s.events.Add() {
		if err := stream.Send(e); err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
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
	if err := s.daemon.Watcher.Sync(); err != nil {
		return pb.NewError(err), nil
	}

	return pb.Success, nil
}

func (s *Server) RestartPeer(ctx context.Context, params *pb.RestartPeerParams) (*pb.Error, error) {
	p, pbErr, err := s.findPeer(params.Intf, params.Peer)
	if pbErr != nil || err != nil {
		return pbErr, err
	}

	ip := s.daemon.EndpointDiscovery.Peers[p]

	ip.Restart()

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
