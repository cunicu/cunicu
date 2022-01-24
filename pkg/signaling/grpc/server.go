package grpc

import (
	"context"
	"io"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Server struct {
	*grpc.Server
	pb.SignalingServer

	topics     map[crypto.Key]*topic
	topicsLock sync.Mutex

	logger *zap.Logger
}

func NewServer() *Server {
	logger := zap.L().Named("server")

	s := &Server{
		Server: grpc.NewServer(),
		logger: logger,
		topics: map[crypto.Key]*topic{},
	}

	pb.RegisterSignalingServer(s, s)

	return s
}

func (s *Server) SubscribeOffers(params *pb.SubscribeOffersParams, stream pb.Signaling_SubscribeOffersServer) error {
	sk := (*crypto.Key)(params.SharedKey)
	top := s.getTopic(sk)

	ch := top.Subscribe()
	for o := range ch {
		err := stream.Send(o)
		if err != nil && err != io.EOF {
			s.logger.Error("Failed to receive offer", zap.Error(err))
		}
	}
	top.Unsubscribe(ch)

	return nil
}

func (s *Server) PublishOffer(ctx context.Context, params *pb.PublishOffersParams) (*pb.Error, error) {
	sk := (*crypto.Key)(params.SharedKey)
	top := s.getTopic(sk)

	top.Publish(params.Offer)

	return pb.Success, nil
}

func (s *Server) getTopic(sk *crypto.Key) *topic {
	s.topicsLock.Lock()
	defer s.topicsLock.Unlock()

	top, ok := s.topics[*sk]
	if ok {
		return top
	} else {
		top := newTopic()
		s.topics[*sk] = top

		return top
	}
}
