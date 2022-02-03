package grpc

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Server struct {
	topicRegistry

	*grpc.Server
	pb.SignalingServer

	logger *zap.Logger
}

func NewServer() *Server {
	logger := zap.L().Named("server")

	s := &Server{
		topicRegistry: topicRegistry{
			topics: map[crypto.Key]*topic{},
		},
		Server: grpc.NewServer(),
		logger: logger,
	}

	pb.RegisterSignalingServer(s, s)

	return s
}

func (s *Server) Subscribe(params *pb.SubscribeParams, stream pb.Signaling_SubscribeServer) error {
	pk, err := crypto.ParseKeyBytes(params.Key)
	if err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	top := s.getTopic(&pk)

	ch := top.Subscribe()
	defer top.Unsubscribe(ch)

	for env := range ch {
		err := stream.Send(env)
		if err != nil && err != io.EOF {
			s.logger.Error("Failed to receive offer", zap.Error(err))
		}
	}

	return nil
}

func (s *Server) Publish(ctx context.Context, env *pb.SignalingEnvelope) (*pb.Error, error) {
	pk, err := crypto.ParseKeyBytes(env.Receipient)
	if err != nil {
		return nil, fmt.Errorf("invalid key: %w", err)
	}

	s.getTopic(&pk).Publish(env)

	return pb.Success, nil
}
