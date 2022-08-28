package rpc

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/util/buildinfo"

	wice "riasc.eu/wice/pkg"
	proto "riasc.eu/wice/pkg/proto"
	rpcproto "riasc.eu/wice/pkg/proto/rpc"
)

type DaemonServer struct {
	rpcproto.UnimplementedDaemonServer

	*Server
	*wice.Daemon
}

func NewDaemonServer(s *Server, d *wice.Daemon) *DaemonServer {
	ds := &DaemonServer{
		Server: s,
		Daemon: d,
	}

	rpcproto.RegisterDaemonServer(s.grpc, ds)

	return ds
}

func (s *DaemonServer) StreamEvents(params *proto.Empty, stream rpcproto.Daemon_StreamEventsServer) error {

	// Send initial connection state of all peers
	if s.epdisc != nil {
		s.epdisc.SendConnectionStates(stream)
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

func (s *DaemonServer) GetBuildInfo(context.Context, *proto.Empty) (*proto.BuildInfo, error) {
	return buildinfo.BuildInfo(), nil
}

func (s *DaemonServer) UnWait(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	err := status.Error(codes.AlreadyExists, "RPC socket has already been unwaited")

	s.waitOnce.Do(func() {
		s.waitGroup.Done()
		err = nil
	})

	return &proto.Empty{}, err
}

func (s *DaemonServer) Stop(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	s.Daemon.Stop()

	return &proto.Empty{}, nil
}

func (s *DaemonServer) Restart(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	if util.ReexecSelfSupported {
		s.Daemon.Restart()
	} else {
		return nil, status.Error(codes.Unimplemented, "not supported on this platform")
	}

	return nil, nil
}
