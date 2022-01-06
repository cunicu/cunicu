package socket

import (
	"context"

	"riasc.eu/wice/pkg/pb"
)

func (s *Server) GetStatus(ctx context.Context, _ *pb.Void) (*pb.Status, error) {
	return &pb.Status{}, nil
}

func (s *Server) StreamEvents(params *pb.StreamEventsParams, stream pb.Socket_StreamEventsServer) error {
	ch := make(chan *pb.Event, 100)

	s.eventListenersLock.Lock()
	s.eventListeners[ch] = nil
	s.eventListenersLock.Unlock()

	for evt := range ch {
		stream.Send(evt)
	}

	return nil
}

func (s *Server) UnWait(ctx context.Context, params *pb.UnWaitParams) (*pb.Error, error) {
	var e = &pb.Error{
		Code:    pb.Error_EALREADY,
		Message: "already unwaited",
	}

	s.waitOnce.Do(func() {
		s.logger.Info("Control socket un-waited")
		s.waitGroup.Done()
		e = pb.Success
	})

	return e, nil
}

func (s *Server) Shutdown(ctx context.Context, params *pb.ShutdownParams) (*pb.Error, error) {
	// Notify main loop
	s.Requests <- params

	return pb.Success, nil
}

func (s *Server) SyncInterfaces(ctx context.Context, params *pb.SyncInterfaceParams) (*pb.Error, error) {
	// Notify main loop
	s.Requests <- params

	return pb.Success, nil
}

func (s *Server) SyncConfig(ctx context.Context, params *pb.SyncConfigParams) (*pb.Error, error) {

	return &pb.Error{
		Code:    pb.Error_ENOTSUP,
		Message: "not implemented yet",
	}, nil
}
