package rpc

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	wice "riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/util"

	"net"
)

type Server struct {
	daemon    *DaemonServer
	epice     *EndpointDiscoveryServer
	watcher   *WatcherServer
	signaling *SignalingServer

	grpc *grpc.Server

	waitGroup sync.WaitGroup
	waitOnce  sync.Once

	events *util.FanOut[*pb.Event]

	logger *zap.Logger
}

func NewServer(d *wice.Daemon, socket string) (*Server, error) {
	s := &Server{
		events: util.NewFanOut[*pb.Event](1),
		logger: zap.L().Named("rpc.server"),
	}

	s.waitGroup.Add(1)

	s.grpc = grpc.NewServer()

	// Register services
	s.daemon = NewDaemonServer(s, d)
	s.watcher = NewWatcherServer(s, d.Watcher)
	s.signaling = NewSignalingServer(s, d.Backend)

	if d.EPDisc != nil {
		s.epice = NewEndpointDiscoveryServer(s, d.EPDisc)
	}

	// Remove old unix sockets
	if err := os.RemoveAll(socket); err != nil {
		return nil, fmt.Errorf("failed to remove old socket: %w", err)
	}

	l, err := net.Listen("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to listen at %s: %w", socket, err)
	}

	go s.grpc.Serve(l)

	return s, nil
}

func (s *Server) Wait() {
	s.logger.Info("Wait for control socket connection")

	s.waitGroup.Wait()

	s.logger.Info("Control socket un-waited")
}

func (s *Server) Close() error {
	s.events.Close()
	s.grpc.GracefulStop()

	return nil
}
