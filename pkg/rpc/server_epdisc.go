// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/proto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

type EndpointDiscoveryServer struct {
	rpcproto.UnimplementedEndpointDiscoverySocketServer

	*Server
}

func NewEndpointDiscoveryServer(s *Server) *EndpointDiscoveryServer {
	eps := &EndpointDiscoveryServer{
		Server: s,
	}

	rpcproto.RegisterEndpointDiscoverySocketServer(s.grpc, eps)

	return eps
}

func (s *EndpointDiscoveryServer) RestartPeer(_ context.Context, params *rpcproto.RestartPeerParams) (*proto.Empty, error) {
	di := s.daemon.InterfaceByName(params.Intf)
	if di == nil {
		return nil, status.Errorf(codes.NotFound, "unknown interface %s", params.Intf)
	}

	i := epdisc.Get(di)
	if i == nil {
		return nil, status.Errorf(codes.NotFound, "interface %s has endpoint discovery not enabled", params.Intf)
	}

	pk, err := crypto.ParseKeyBytes(params.Peer)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse peer public key: %s", err)
	}

	p := i.PeerByPublicKey(pk)
	if p == nil {
		return nil, status.Errorf(codes.NotFound, "unknown peer %s/%s", params.Intf, pk)
	}

	if err = p.Restart(); err != nil {
		return &proto.Empty{}, status.Errorf(codes.Unknown, "failed to restart peer session: %s", err)
	}

	return &proto.Empty{}, nil
}
