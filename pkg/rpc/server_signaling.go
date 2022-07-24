package rpc

import (
	"context"

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
	// peer, pbErr, err := s.findPeer(params.Intf, params.Peer)
	// if pbErr != nil || err != nil {
	// 	return nil, err
	// }

	return &pb.GetSignalingMessageResp{}, nil
}

func (s *SignalingServer) PutSignalingMessage(ctx context.Context, params *pb.PutSignalingMessageParams) (*pb.Error, error) {

	return pb.Success, nil
}

func (s *SignalingServer) Subscribe(*pb.SubscribeParams, pb.Signaling_SubscribeServer) error {
	return nil
}

func (s *SignalingServer) Publish(context.Context, *signaling.Envelope) (*pb.Error, error) {
	return pb.Success, nil
}
