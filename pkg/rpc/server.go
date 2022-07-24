package rpc

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/pb"

	"net"
)

type Server struct {
	daemon    *DaemonServer
	ep        *EndpointDiscoveryServer
	watcher   *WatcherServer
	signaling *SignalingServer

	grpc *grpc.Server

	waitGroup sync.WaitGroup
	waitOnce  sync.Once

	events *util.FanOut[*pb.Event]

	logger *zap.Logger
}

func NewServer(d *pkg.Daemon) (*Server, error) {
	s := &Server{
		events: util.NewFanOut[*pb.Event](0),
		logger: zap.L().Named("socket.server"),
	}

	s.waitGroup.Add(1)

	s.grpc = grpc.NewServer(
		grpc.UnaryInterceptor(s.UnaryInterceptor),
		grpc.StreamInterceptor(s.StreamInterceptor),
	)

	// Register services
	s.daemon = NewDaemonServer(s, d)
	s.watcher = NewWatcherServer(s, d.Watcher)
	s.signaling = NewSignalingServer(s, d.Backend)

	if d.EndpointDiscovery != nil {
		s.ep = NewEndpointDiscoveryServer(s, d.EndpointDiscovery)
	}

	return s, nil
}

func (s *Server) Listen(network string, address string) error {
	// Remove old unix sockets
	if network == "unix" {
		if err := os.RemoveAll(address); err != nil {
			return fmt.Errorf("failed to remove old socket: %w", err)
		}
	}

	l, err := net.Listen(network, address)
	if err != nil {
		return fmt.Errorf("failed to listen on %s %s", network, address)
	}

	go s.grpc.Serve(l)

	return nil
}

func (s *Server) Wait() {
	s.logger.Info("Wait for control socket connection")

	s.waitGroup.Wait()
}

func (s *Server) Close() error {
	s.grpc.GracefulStop()

	return nil
}
