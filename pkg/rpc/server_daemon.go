package rpc

import (
	"context"
	"fmt"

	wice "riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/pb"
)

type DaemonServer struct {
	pb.UnimplementedSocketServer

	*Server
	*wice.Daemon
}

func NewDaemonServer(s *Server, d *wice.Daemon) *DaemonServer {
	ds := &DaemonServer{
		Server: s,
		Daemon: d,
	}

	pb.RegisterSocketServer(s.grpc, ds)

	return ds
}

func (s *DaemonServer) StreamEvents(params *pb.StreamEventsParams, stream pb.Socket_StreamEventsServer) error {

	// Send initial connection state of all peers
	if s.ep != nil {
		s.ep.SendConnectionStates(stream)
	}

	for e := range s.events.Add() {
		if err := stream.Send(e); err != nil {
			return fmt.Errorf("failed to send event: %w", err)
		}
	}

	return nil
}

func (s *DaemonServer) UnWait(ctx context.Context, params *pb.UnWaitParams) (*pb.Error, error) {
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

func (s *DaemonServer) Stop(ctx context.Context, params *pb.StopParams) (*pb.Error, error) {
	if err := s.Daemon.Close(); err != nil {
		return pb.NewError(err), nil
	}

	return pb.Success, nil
}
