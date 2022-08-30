package rpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/proto"
	coreproto "riasc.eu/wice/pkg/proto/core"
	rpcproto "riasc.eu/wice/pkg/proto/rpc"
	"riasc.eu/wice/pkg/watcher"
)

type WatcherServer struct {
	rpcproto.UnimplementedWatcherServer

	*Server
	*watcher.Watcher
}

func NewWatcherServer(s *Server, w *watcher.Watcher) *WatcherServer {
	ws := &WatcherServer{
		Server:  s,
		Watcher: w,
	}

	rpcproto.RegisterWatcherServer(s.grpc, ws)

	w.OnAll(s)

	return ws
}

func (s *WatcherServer) GetStatus(ctx context.Context, p *rpcproto.StatusParams) (*rpcproto.StatusResp, error) {
	var err error
	var pk crypto.Key

	if p.Peer != nil {
		if pk, err = crypto.ParseKeyBytes(p.Peer); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid peer key")
		}
	}

	qis := []*coreproto.Interface{}
	s.ForEachInterface(func(ci *core.Interface) error {
		if p.Intf == "" || ci.Name() == p.Intf {
			qis = append(qis, ci.MarshalWithPeers(func(cp *core.Peer) *coreproto.Peer {
				if pk.IsSet() && pk != cp.PublicKey() {
					return nil
				}

				qp := cp.Marshal()

				if s.epdisc != nil {
					qp.Ice = s.epdisc.PeerStatus(cp)
				}

				return qp
			}))
		}

		return nil
	})

	// Check if filters matched anything
	if p.Intf != "" && len(qis) == 0 {
		return nil, status.Errorf(codes.NotFound, "no such interface '%s'", p.Intf)
	} else if pk.IsSet() && len(qis[0].Peers) == 0 {
		return nil, status.Errorf(codes.NotFound, "no such peer '%s' for interface '%s'", pk, p.Intf)
	}

	return &rpcproto.StatusResp{
		Interfaces: qis,
	}, nil
}

func (s *WatcherServer) Sync(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	if err := s.Watcher.Sync(); err != nil {
		return &proto.Empty{}, status.Errorf(codes.Unknown, "failed to sync: %s", err)
	}

	return &proto.Empty{}, nil
}
