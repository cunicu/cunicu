package socket

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"riasc.eu/wice/pkg/pb"

	"net"
	"os"
)

type Server struct {
	pb.SocketServer

	listener net.Listener
	grpc     *grpc.Server

	eventListeners     map[chan *pb.Event]interface{}
	eventListenersLock sync.Mutex

	waitGroup sync.WaitGroup
	waitOnce  sync.Once

	logger *log.Entry
}

func Listen(network string, address string, wait bool) (*Server, error) {
	// Remove old unix sockets
	if network == "unix" {
		if err := os.RemoveAll(address); err != nil {
			log.Fatal(err)
		}
	}

	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	s := &Server{
		listener:       l,
		logger:         log.WithField("logger", "socket"),
		grpc:           grpc.NewServer(),
		eventListeners: map[chan *pb.Event]interface{}{},
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

func (s *Server) GetStatus(ctx context.Context, _ *pb.Void) (*pb.Status, error) {
	return &pb.Status{}, nil
}

func (s *Server) StreamEvents(_ *pb.Void, stream pb.Socket_StreamEventsServer) error {
	ch := make(chan *pb.Event, 100)

	s.eventListenersLock.Lock()
	s.eventListeners[ch] = nil
	s.eventListenersLock.Unlock()

	for evt := range ch {
		stream.Send(evt)
	}

	return nil
}

func (s *Server) UnWait(context.Context, *pb.Void) (*pb.Error, error) {
	var e = &pb.Error{
		Ok:    false,
		Error: "already unwaited",
	}

	s.waitOnce.Do(func() {
		s.logger.Info("Control socket un-waited")
		s.waitGroup.Done()
		e = &pb.Ok
	})

	return e, nil
}
