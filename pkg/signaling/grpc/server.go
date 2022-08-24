package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func NewServer(opts ...grpc.ServerOption) *Server {
	logger := zap.L().Named("server")

	if fn := os.Getenv("SSLKEYLOGFILE"); fn != "" {
		wr, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			logger.Fatal("Failed to open SSL keylog file", zap.Error(err))
		}

		opts = slices.Clone(opts)
		opts = append(opts, grpc.Creds(
			credentials.NewTLS(&tls.Config{
				KeyLogWriter: wr,
			}),
		))
	}

	s := &Server{
		topicRegistry: topicRegistry{
			topics: map[crypto.Key]*topic{},
		},
		Server: grpc.NewServer(opts...),
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

	// We send an empty envelope to signal the subscriber that the subscription
	// has been created. This avoids a race between Subscribe() & Publish() from the
	// clients view-point.
	if err := stream.Send(&pb.SignalingEnvelope{}); err != nil {
		s.logger.Error("Failed to send sync envelope", zap.Error(err))
	}

out:
	for {
		select {
		case env := <-ch:
			if err := stream.Send(env); err == io.EOF {
				break out
			} else if err != nil {
				s.logger.Error("Failed to send envelope", zap.Error(err))
			}

		case <-stream.Context().Done():
			break out
		}
	}

	return nil
}

func (s *Server) Publish(ctx context.Context, env *signaling.Envelope) (*pb.Empty, error) {
	var err error
	var pkRecipient, pkSender crypto.Key

	if pkRecipient, err = crypto.ParseKeyBytes(env.Recipient); err != nil {
		return &pb.Empty{}, fmt.Errorf("invalid recipient key: %w", err)
	}

	if pkSender, err = crypto.ParseKeyBytes(env.Sender); err != nil {
		return &pb.Empty{}, fmt.Errorf("invalid sender key: %w", err)
	}

	t := s.getTopic(&pkRecipient)

	t.Publish(env)

	s.logger.Debug("Published envelope",
		zap.Any("recipient", pkRecipient),
		zap.Any("sender", pkSender))

	return &pb.Empty{}, nil
}

func (s *Server) GracefulStop() {
	// Close all subscription streams
	s.topicRegistry.Close()

	s.Server.GracefulStop()
}
