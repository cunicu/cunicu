// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "cunicu.li/cunicu/pkg/proto"
	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
	signalingproto "cunicu.li/cunicu/pkg/proto/signaling"
	"cunicu.li/cunicu/pkg/signaling"
	"cunicu.li/cunicu/pkg/signaling/grpc"
)

type SignalingServer struct {
	rpcproto.UnimplementedSignalingServer

	*Server
	*grpc.Backend
}

func NewSignalingServer(s *Server, b *signaling.MultiBackend) *SignalingServer {
	gb, ok := b.ByType(signalingproto.BackendType_GRPC).(*grpc.Backend)
	if !ok {
		return nil
	}

	ss := &SignalingServer{
		Server:  s,
		Backend: gb,
	}

	rpcproto.RegisterSignalingServer(s.grpc, ss)

	return ss
}

func (s *SignalingServer) GetSignalingMessage(_ context.Context, _ *rpcproto.GetSignalingMessageParams) (*rpcproto.GetSignalingMessageResp, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *SignalingServer) PutSignalingMessage(_ context.Context, _ *rpcproto.PutSignalingMessageParams) (*proto.Empty, error) {
	return &proto.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}
