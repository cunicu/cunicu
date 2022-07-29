package grpc

import (
	"context"
	"fmt"
	"io"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

type Server struct {
	topicRegistry

	*grpc.Server
	pb.SignalingServer

	logger *zap.Logger
}

func NewServer(opt ...grpc.ServerOption) *Server {
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
		if err := stream.Send(env); err != nil && err != io.EOF {
			s.logger.Error("Failed to receive envelope", zap.Error(err))
		}
	}

	return nil
}

func (s *Server) Publish(ctx context.Context, env *signaling.Envelope) (*pb.Error, error) {
	var err error
	var pkRecipient, pkSender crypto.Key

	if pkRecipient, err = crypto.ParseKeyBytes(env.Recipient); err != nil {
		return nil, fmt.Errorf("invalid recipient key: %w", err)
	}

	if pkSender, err = crypto.ParseKeyBytes(env.Sender); err != nil {
		return nil, fmt.Errorf("invalid sender key: %w", err)
	}

	t := s.getTopic(&pkRecipient)

	// Publishing a message to a topic in which we are the only subscriber is
	// meaningless as the message will have no audience.
	t.WaitForSubs(1)

	t.Publish(env)

	s.logger.Debug("Published envelope",
		zap.Any("recipient", pkRecipient),
		zap.Any("sender", pkSender),
		zap.Int("num_subs", len(t.subs)))

	return pb.Success, nil
}
