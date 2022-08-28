package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/grpc"

	proto "riasc.eu/wice/pkg/proto"
	rpcproto "riasc.eu/wice/pkg/proto/rpc"
	signalingproto "riasc.eu/wice/pkg/proto/signaling"
)

type SignalingServer struct {
	rpcproto.UnimplementedSignalingServer

	*Server
	*grpc.Backend
}

func NewSignalingServer(s *Server, b *signaling.MultiBackend) *SignalingServer {
	gb := b.ByType(signalingproto.BackendType_GRPC).(*grpc.Backend)

	ss := &SignalingServer{
		Server:  s,
		Backend: gb,
	}

	rpcproto.RegisterSignalingServer(s.grpc, ss)

	return ss
}

func (s *SignalingServer) GetSignalingMessage(ctx context.Context, params *rpcproto.GetSignalingMessageParams) (*rpcproto.GetSignalingMessageResp, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *SignalingServer) PutSignalingMessage(ctx context.Context, params *rpcproto.PutSignalingMessageParams) (*proto.Empty, error) {
	return &proto.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}
