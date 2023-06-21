// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/stv0g/cunicu/pkg/proto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/signaling/grpc"
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
