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

func (s *DaemonServer) StreamEvents(params *pb.Empty, stream pb.Socket_StreamEventsServer) error {

	// Send initial connection state of all peers
	if s.epice != nil {
		s.epice.SendConnectionStates(stream)
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

func (s *DaemonServer) UnWait(ctx context.Context, params *pb.Empty) (*pb.Empty, error) {
	err := status.Error(codes.AlreadyExists, "RPC socket has already been unwaited")

	s.waitOnce.Do(func() {
		s.waitGroup.Done()
		err = nil
	})

	return &pb.Empty{}, err
}

func (s *DaemonServer) Stop(ctx context.Context, params *pb.Empty) (*pb.Empty, error) {
	s.Daemon.Stop()

	return &pb.Empty{}, nil
}
