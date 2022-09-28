package rpc

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"

	cunicu "github.com/stv0g/cunicu/pkg"
	proto "github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

type DaemonServer struct {
	rpcproto.UnimplementedDaemonServer

	*Server
	*cunicu.Daemon
}

func NewDaemonServer(s *Server, d *cunicu.Daemon) *DaemonServer {
	ds := &DaemonServer{
		Server: s,
		Daemon: d,
	}

	rpcproto.RegisterDaemonServer(s.grpc, ds)

	return ds
}

func (s *DaemonServer) StreamEvents(params *proto.Empty, stream rpcproto.Daemon_StreamEventsServer) error {

	// Send initial connection state of all peers
	if s.epdisc != nil {
		s.epdisc.SendConnectionStates(stream)
	}

	events := s.events.Add()
	defer s.events.Remove(events)

out:
	for {
		select {
		case event := <-events:
			if err := stream.Send(event); err == io.EOF {
				break out
			} else if err != nil {
				return fmt.Errorf("failed to send event: %w", err)
			}

		case <-stream.Context().Done():
			break out
		}
	}

	return nil
}

func (s *DaemonServer) GetBuildInfo(context.Context, *proto.Empty) (*proto.BuildInfo, error) {
	return buildinfo.BuildInfo(), nil
}

func (s *DaemonServer) UnWait(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	err := status.Error(codes.AlreadyExists, "RPC socket has already been unwaited")

	s.waitOnce.Do(func() {
		s.waitGroup.Done()
		err = nil
	})

	return &proto.Empty{}, err
}

func (s *DaemonServer) Stop(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	s.Daemon.Stop()

	return &proto.Empty{}, nil
}

func (s *DaemonServer) Restart(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	if util.ReexecSelfSupported {
		s.Daemon.Restart()
	} else {
		return nil, status.Error(codes.Unimplemented, "not supported on this platform")
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) Sync(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	if err := s.Daemon.Sync(); err != nil {
		return &proto.Empty{}, status.Errorf(codes.Unknown, "failed to sync: %s", err)
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) GetStatus(ctx context.Context, p *rpcproto.GetStatusParams) (*rpcproto.GetStatusResp, error) {
	var err error
	var pk crypto.Key

	if p.Peer != nil {
		if pk, err = crypto.ParseKeyBytes(p.Peer); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid peer key")
		}
	}

	qis := []*coreproto.Interface{}
	s.daemon.ForEachInterface(func(ci *core.Interface) error {
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
	if p.Interface != "" && len(qis) == 0 {
		return nil, status.Errorf(codes.NotFound, "no such interface '%s'", p.Interface)
	} else if pk.IsSet() && len(qis[0].Peers) == 0 {
		return nil, status.Errorf(codes.NotFound, "no such peer '%s' for interface '%s'", pk, p.Interface)
	}

	return &rpcproto.GetStatusResp{
		Interfaces: qis,
	}, nil
}

func (s *DaemonServer) SetConfig(ctx context.Context, p *rpcproto.SetConfigParams) (*proto.Empty, error) {
	errs := []error{}
	settings := map[string]any{}

	for key, value := range p.Settings {
		switch key {
		case "log.verbosity":
			level, err := strconv.Atoi(value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid level: %w", err))
				break
			} else if level > 10 || level < 0 {
				errs = append(errs, fmt.Errorf("invalid level (must be between 0 and 10 inclusive)"))
				break
			}

			log.Verbosity.SetLevel(level)

		case "log.severity":
			level, err := zapcore.ParseLevel(value)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid level: %w", err))
				break
			} else if level < zapcore.DebugLevel || level > zapcore.FatalLevel {
				errs = append(errs, fmt.Errorf("invalid level"))
				break
			}

			log.Severity.SetLevel(level)

		default:
			settings[key] = value
		}
	}

	if err := s.Config.Update(settings); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		errstrs := []string{}
		for _, err := range errs {
			errstrs = append(errstrs, err.Error())
		}

		return nil, status.Error(codes.InvalidArgument, strings.Join(errstrs, ", "))
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) GetConfig(ctx context.Context, p *rpcproto.GetConfigParams) (*rpcproto.GetConfigResp, error) {
	settings := map[string]string{}

	match := func(key string) bool {
		return p.KeyFilter == "" || strings.HasPrefix(key, p.KeyFilter)
	}

	if match("log.verbosity") {
		settings["log.verbosity"] = fmt.Sprint(log.Verbosity.Level())
	}

	if match("log.severity") {
		settings["log.severity"] = log.Severity.String()
	}

	for key, value := range s.Config.All() {
		if match(key) {
			settings[key] = fmt.Sprintf("%v", value)
		}
	}

	return &rpcproto.GetConfigResp{
		Settings: settings,
	}, nil
}

func (s *DaemonServer) ReloadConfig(ctx context.Context, params *proto.Empty) (*proto.Empty, error) {
	if err := s.Config.Reload(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to reload configuration: %s", err)
	}

	return &proto.Empty{}, nil
}
