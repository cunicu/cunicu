package socket

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
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

	events *util.FanOut[*pb.Event]

	logger *zap.Logger
}

func Listen(network string, address string) (*Server, error) {
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
		listener: l,
		events:   util.NewFanOut[*pb.Event](0),
		logger:   logger,
		grpc:     grpc.NewServer(),
	}

	s.waitGroup.Add(1)

	go s.grpc.Serve(l)

	return s, nil
}

func (s *Server) Wait() {
	s.logger.Info("Wait for control socket connection")

	s.waitGroup.Wait()
}

func (s *Server) RegisterDaemon(d *pkg.Daemon) {
	s.daemon = d

	pb.RegisterSocketServer(s.grpc, s)

	d.Watcher.RegisterAll(s)

	if d.EndpointDiscovery != nil {
		d.EndpointDiscovery.OnConnectionStateChange(s)
	}
}

func (s *Server) findPeer(intfName string, peerPK []byte) (*core.Peer, *pb.Error, error) {
	intf := s.daemon.Watcher.Interfaces.ByName(intfName)
	if intf == nil {
		return nil, &pb.Error{
			Code:    pb.Error_ENOENT,
			Message: "Interface not found",
		}, nil
	}

	pk, err := crypto.ParseKeyBytes(peerPK)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid key: %w", err)
	}

	peer, ok := intf.Peers[pk]
	if !ok {
		return nil, &pb.Error{
			Code:    pb.Error_ENOENT,
			Message: "Peer not found",
		}, nil
	}

	return peer, nil, nil
}
