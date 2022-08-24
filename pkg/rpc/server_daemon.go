package rpc

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	events := s.events.Add()
	defer s.events.Remove(events)

out:
	for {
		select {
		case event := <-events:
			if err := stream.Send(event); err == io.EOF {
				break out
			} else if err != nil {
				return fmt.Errorf("failed to send event: %w", err)
			}

		case <-stream.Context().Done():
			break out
		}
	}

	return nil
}

func (s *DaemonServer) UnWait(ctx context.Context, params *pb.UnWaitParams) (*pb.Empty, error) {
	err := status.Error(codes.AlreadyExists, "RPC socket has already been unwaited")

	s.waitOnce.Do(func() {
		s.logger.Info("Control socket un-waited")
		s.waitGroup.Done()
		err = nil
	})

	return &pb.Empty{}, err
}

func (s *DaemonServer) Stop(ctx context.Context, params *pb.StopParams) (*pb.Empty, error) {
	if err := s.Daemon.Close(); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to stop daemon: %s", err)
	}

	return &pb.Empty{}, nil
}
