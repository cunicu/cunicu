package rpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func (s *Server) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to get peer from context")
	}

	s.logger.Info("Intercepted unary call",
		zap.String("client_ip", p.Addr.String()),
		zap.String("method", info.FullMethod),
	)

	return handler(ctx, req)
}

func (s *Server) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	p, ok := peer.FromContext(ss.Context())
	if !ok {
		return errors.New("failed to get peer from context")
	}

	s.logger.Info("Intercepted stream",
		zap.String("client_ip", p.Addr.String()),
		zap.String("method", info.FullMethod),
	)

	return handler(srv, ss)
}
