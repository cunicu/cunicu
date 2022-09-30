package rpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/util"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

type Server struct {
	daemon    *DaemonServer
	epdisc    *EndpointDiscoveryServer
	signaling *SignalingServer

	grpc *grpc.Server

	waitGroup sync.WaitGroup
	waitOnce  sync.Once

	events *util.FanOut[*rpcproto.Event]

	logger *zap.Logger
}

func NewServer(d *daemon.Daemon, socket string) (*Server, error) {
	s := &Server{
		events: util.NewFanOut[*rpcproto.Event](1),
		logger: zap.L().Named("rpc.server"),
	}

	s.waitGroup.Add(1)

	s.grpc = grpc.NewServer(grpc.UnaryInterceptor(s.unaryInterceptor))

	// Register services
	s.daemon = NewDaemonServer(s, d)
	s.signaling = NewSignalingServer(s, d.Backend)
	s.epdisc = NewEndpointDiscoveryServer(s)

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

func (s *Server) unaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		s.logger.Error("Failed to handle RPC request",
			zap.Error(err),
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
		)
	} else {
		s.logger.Debug("Handling RPC request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req),
			zap.Any("response", resp),
		)
	}

	return
}
