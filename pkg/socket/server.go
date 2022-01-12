package socket

import (
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/pb"

	"net"
	"os"
)

type Server struct {
	pb.SocketServer

	daemon *pkg.Daemon

	listener net.Listener
	grpc     *grpc.Server

	waitGroup sync.WaitGroup
	waitOnce  sync.Once

	logger *zap.Logger
}

func Listen(network string, address string, wait bool, daemon *pkg.Daemon) (*Server, error) {
	logger := zap.L().Named("socket.server")
	// Remove old unix sockets
	if network == "unix" {
		if err := os.RemoveAll(address); err != nil {
			logger.Fatal("Failed to remove old socket", zap.Error(err))
		}
	}

	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	s := &Server{
		daemon:   daemon,
		listener: l,
		logger:   logger,
		grpc:     grpc.NewServer(),
	}

	pb.RegisterSocketServer(s.grpc, s)

	go s.grpc.Serve(l)

	s.waitGroup.Add(1)
	if wait {
		s.logger.Info("Wait for control socket connection")

		s.waitGroup.Wait()
	}

	return s, nil
}
