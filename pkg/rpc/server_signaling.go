package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/signaling/grpc"
)

type SignalingServer struct {
	pb.UnimplementedSignalingSocketServer
	pb.UnimplementedSignalingServer

	*Server
	*grpc.Backend
}

func NewSignalingServer(s *Server, b *signaling.MultiBackend) *SignalingServer {
	gb := b.ByType(pb.BackendType_GRPC).(*grpc.Backend)

	ss := &SignalingServer{
		Server:  s,
		Backend: gb,
	}

	pb.RegisterSignalingSocketServer(s.grpc, ss)
	pb.RegisterSignalingServer(s.grpc, ss)

	return ss
}

func (s *SignalingServer) GetSignalingMessage(ctx context.Context, params *pb.GetSignalingMessageParams) (*pb.GetSignalingMessageResp, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *SignalingServer) PutSignalingMessage(ctx context.Context, params *pb.PutSignalingMessageParams) (*pb.Empty, error) {
	return &pb.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *SignalingServer) Subscribe(*pb.SubscribeParams, pb.Signaling_SubscribeServer) error {
	return status.Error(codes.Unimplemented, "not implemented yet")
}

func (s *SignalingServer) Publish(context.Context, *signaling.Envelope) (*pb.Empty, error) {
	return &pb.Empty{}, status.Error(codes.Unimplemented, "not implemented yet")
}
