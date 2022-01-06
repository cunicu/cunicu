package socket

import (
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"riasc.eu/wice/pkg/pb"

	"net"
	"os"
)

type Server struct {
	pb.SocketServer

	listener net.Listener
	grpc     *grpc.Server

	Requests chan interface{}

	eventListeners     map[chan *pb.Event]interface{}
	eventListenersLock sync.Mutex

	waitGroup sync.WaitGroup
	waitOnce  sync.Once

	logger *zap.Logger
}

func Listen(network string, address string, wait bool) (*Server, error) {
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
		listener:       l,
		logger:         logger,
		grpc:           grpc.NewServer(),
		eventListeners: map[chan *pb.Event]interface{}{},
		Requests:       make(chan interface{}),
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

func (s *Server) BroadcastEvent(e *pb.Event) error {
	if e.Time == nil {
		e.Time = pb.TimeNow()
	}

	s.eventListenersLock.Lock()
	for ch := range s.eventListeners {
		ch <- e
	}
	s.eventListenersLock.Unlock()

	e.Log(s.logger, "Broadcasted event")

	return nil
}
