package grpc

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/proto"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"

	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
)

type Server struct {
	signalingproto.UnimplementedSignalingServer

	topicRegistry

	*grpc.Server

	logger *zap.Logger
}

func NewSignalingServer(opts ...grpc.ServerOption) *Server {
	logger := zap.L().Named("grpc.server")

	s := &Server{
		topicRegistry: topicRegistry{
			topics: map[crypto.Key]*topic{},
		},
		Server: grpc.NewServer(opts...),
		logger: logger,
	}

	signalingproto.RegisterSignalingServer(s, s)

	return s
}

func NewServer(opts ...grpc.ServerOption) (*grpc.Server, error) {
	if fn := os.Getenv("SSLKEYLOGFILE"); fn != "" {
		//#nosec G304 -- Filename is only controlled via env var
		wr, err := os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("failed to open SSL keylog file: %w", err)
		}

		opts = slices.Clone(opts)
		opts = append(opts,
			grpc.Creds(
				credentials.NewTLS(&tls.Config{
					MinVersion:   tls.VersionTLS13,
					KeyLogWriter: wr,
				}),
			),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: 5 * time.Second,
			}),
		)
	}

	return grpc.NewServer(opts...), nil
}

func (s *Server) Subscribe(params *signalingproto.SubscribeParams, stream signalingproto.Signaling_SubscribeServer) error {
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
	if err := stream.Send(&signalingproto.Envelope{}); err != nil {
		s.logger.Error("Failed to send sync envelope", zap.Error(err))
	}

out:
	for {
		select {
		case env, ok := <-ch:
			if !ok {
				break out
			}

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

func (s *Server) Publish(ctx context.Context, env *signaling.Envelope) (*proto.Empty, error) {
	var err error
	var pkRecipient, pkSender crypto.Key

	if pkRecipient, err = crypto.ParseKeyBytes(env.Recipient); err != nil {
		return &proto.Empty{}, fmt.Errorf("invalid recipient key: %w", err)
	}

	if pkSender, err = crypto.ParseKeyBytes(env.Sender); err != nil {
		return &proto.Empty{}, fmt.Errorf("invalid sender key: %w", err)
	}

	t := s.getTopic(&pkRecipient)

	t.Publish(env)

	s.logger.Debug("Published envelope",
		zap.Any("recipient", pkRecipient),
		zap.Any("sender", pkSender))

	return &proto.Empty{}, nil
}

func (s *Server) Close() error {
	if err := s.topicRegistry.Close(); err != nil {
		return err
	}

	s.Server.GracefulStop()

	return nil
}

func (s *Server) GetBuildInfo(context.Context, *proto.Empty) (*proto.BuildInfo, error) {
	return buildinfo.BuildInfo(), nil
}
