// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rpc

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cunicu.li/cunicu/pkg/buildinfo"
	"cunicu.li/cunicu/pkg/crypto"
	"cunicu.li/cunicu/pkg/daemon"
	"cunicu.li/cunicu/pkg/daemon/feature/epdisc"
	osx "cunicu.li/cunicu/pkg/os"
	"cunicu.li/cunicu/pkg/proto"
	coreproto "cunicu.li/cunicu/pkg/proto/core"
	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
	slicesx "cunicu.li/cunicu/pkg/types/slices"
)

var errNoSettingChanged = errors.New("no setting was changed")

type DaemonServer struct {
	rpcproto.UnimplementedDaemonServer

	*Server
	*daemon.Daemon
}

func NewDaemonServer(s *Server, d *daemon.Daemon) *DaemonServer {
	ds := &DaemonServer{
		Server: s,
		Daemon: d,
	}

	rpcproto.RegisterDaemonServer(s.grpc, ds)

	d.AddInterfaceHandler(ds)

	return ds
}

func (s *DaemonServer) StreamEvents(_ *proto.Empty, stream rpcproto.Daemon_StreamEventsServer) error {
	// Send initial connection state of all peers
	s.SendPeerStates(stream)

	events := s.events.Add()
	defer s.events.Remove(events)

out:
	for {
		select {
		case event, ok := <-events:
			if !ok {
				break out
			}

			if err := stream.Send(event); errors.Is(err, io.EOF) {
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

func (s *DaemonServer) UnWait(_ context.Context, _ *proto.Empty) (*proto.Empty, error) {
	err := status.Error(codes.AlreadyExists, "RPC socket has already been unwaited")

	s.waitOnce.Do(func() {
		s.waitGroup.Done()

		err = nil
	})

	return &proto.Empty{}, err
}

func (s *DaemonServer) Shutdown(_ context.Context, params *rpcproto.ShutdownParams) (*proto.Empty, error) {
	if params.Restart && !osx.ReexecSelfSupported {
		return nil, status.Error(codes.Unimplemented, "not supported on this platform")
	}

	s.Daemon.Shutdown(params.Restart)

	return &proto.Empty{}, nil
}

func (s *DaemonServer) Sync(_ context.Context, _ *proto.Empty) (*proto.Empty, error) {
	if err := s.Daemon.Sync(); err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to sync: %s", err)
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) GetStatus(_ context.Context, p *rpcproto.GetStatusParams) (*rpcproto.GetStatusResp, error) { //nolint:gocognit
	var (
		err error
		pk  crypto.Key
	)

	if p.Peer != nil {
		if pk, err = crypto.ParseKeyBytes(p.Peer); err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid peer key")
		}
	}

	qis := []*coreproto.Interface{}

	if err := s.daemon.ForEachInterface(func(i *daemon.Interface) error {
		epi := epdisc.Get(i)

		if p.Interface == "" || i.Name() == p.Interface {
			qi := i.MarshalWithPeers(func(cp *daemon.Peer) *coreproto.Peer {
				if pk.IsSet() && pk != cp.PublicKey() {
					return nil
				}

				qp := cp.Marshal()

				if epi != nil {
					if epp, ok := epi.Peers[cp]; ok {
						qp.Ice = epp.Marshal()
						qp.Reachability = epp.Reachability()
					}
				}

				return qp
			})

			if epi != nil {
				qi.Ice = epi.Marshal()
			}

			qis = append(qis, qi)
		}

		return nil
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal interface: %s", err)
	}

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

func (s *DaemonServer) SetConfig(_ context.Context, p *rpcproto.SetConfigParams) (*proto.Empty, error) {
	settings := map[string]any{}

	for key, value := range p.Settings {
		switch {
		case value.Scalar != "":
			settings[key] = value.Scalar
		case len(value.List) > 0:
			settings[key] = value.List
		default:
			settings[key] = nil // Unset
		}
	}

	if changes, err := s.Config.Update(settings); err != nil {
		return nil, decodeError(err)
	} else if len(changes) == 0 {
		return nil, status.Error(codes.InvalidArgument, errNoSettingChanged.Error())
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) GetConfig(_ context.Context, p *rpcproto.GetConfigParams) (*rpcproto.GetConfigResp, error) {
	settings := map[string]*rpcproto.ConfigValue{}

	for key, value := range s.Config.All() {
		if p.KeyFilter != "" && !strings.HasPrefix(key, p.KeyFilter) {
			continue
		}

		str, err := settingToValue(value)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to marshal: %s", err)
		}

		settings[key] = str
	}

	return &rpcproto.GetConfigResp{
		Settings: settings,
	}, nil
}

func (s *DaemonServer) GetCompletion(_ context.Context, params *rpcproto.GetCompletionParams) (*rpcproto.GetCompletionResp, error) {
	var (
		options []string
		flags   cobra.ShellCompDirective
	)

	switch {
	case len(params.Cmd) < 2 || params.Cmd[0] != "cunicu":
		flags = cobra.ShellCompDirectiveError
	case params.Cmd[1] == "config":
		flags = cobra.ShellCompDirectiveNoFileComp
		options = s.getConfigCompletion(params.Cmd[2], params.Args, params.ToComplete)
	default:
		flags = cobra.ShellCompDirectiveNoFileComp
	}

	return &rpcproto.GetCompletionResp{
		Options: options,
		Flags:   int32(flags), //nolint:gosec
	}, nil
}

func (s *DaemonServer) getConfigCompletion(cmd string, args []string, toComplete string) []string {
	var options []string

	if isValueCompletion := len(args) > 0; isValueCompletion {
		if cmd != "set" {
			return nil
		}

		if meta := s.Config.Meta.Lookup(args[0]); meta != nil {
			options = meta.CompletionOptions()
		}
	} else {
		options = s.Config.Meta.Keys()
	}

	if toComplete != "" {
		options = slicesx.Filter(options, func(s string) bool {
			return strings.HasPrefix(s, toComplete)
		})
	}

	return options
}

func (s *DaemonServer) ReloadConfig(_ context.Context, _ *proto.Empty) (*proto.Empty, error) {
	if _, err := s.Config.ReloadAllSources(); err != nil {
		return nil, decodeError(err)
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) AddPeer(_ context.Context, params *rpcproto.AddPeerParams) (*rpcproto.AddPeerResp, error) {
	i := s.InterfaceByName(params.Interface)
	if i == nil {
		return nil, status.Errorf(codes.NotFound, "Interface %s does not exist", params.Interface)
	}

	// Add peer to running daemon
	pk, err := crypto.ParseKeyBytes(params.PublicKey)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid public key")
	}

	if err := i.AddPeer(&wgtypes.PeerConfig{
		PublicKey: wgtypes.Key(pk),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to add peer: %s", err)
	}

	// TODO: Persist new peer in runtime config

	// Create response
	resp := &rpcproto.AddPeerResp{
		Invitation: &rpcproto.Invitation{},
		Interface:  i.Marshal(),
	}

	if community := crypto.Key(i.Settings.Community); community.IsSet() {
		resp.Invitation.Community = community.Bytes()
	}

	// Detect our own endpoint
	if epi := epdisc.Get(i); epi != nil {
		if ep, err := epi.Endpoint(); err != nil {
			s.logger.Warn("Failed to determine our own endpoint address", zap.Error(err))
		} else if ep != nil {
			resp.Invitation.Endpoint = ep.String()

			// Perform reverse lookup of our own endpoint
			if names, err := net.LookupAddr(resp.Invitation.Endpoint); err == nil && len(names) >= 1 {
				epName := names[0]

				// Do not use auto-generated IPv4 rDNS names
				if match, _ := regexp.MatchString(`\d{1,3}-\d{1,3}-\d{1,3}-\d{1,3}`, epName); !match {
					resp.Invitation.Endpoint = fmt.Sprintf("%s:%d", epName, ep.Port)
				}
			}
		}
	}

	return resp, nil
}

func (s *DaemonServer) OnInterfaceAdded(i *daemon.Interface) {
	i.AddPeerStateChangeHandler(s)
}

func (s *DaemonServer) OnInterfaceRemoved(_ *daemon.Interface) {
}

func (s *DaemonServer) SendPeerStates(stream rpcproto.Daemon_StreamEventsServer) {
	if err := s.daemon.ForEachInterface(func(di *daemon.Interface) error {
		if i := epdisc.Get(di); i != nil {
			for _, p := range i.Peers {
				e := &rpcproto.Event{
					Type:      rpcproto.EventType_PEER_STATE_CHANGED,
					Interface: p.Interface.Name(),
					Peer:      p.Peer.PublicKey().Bytes(),
					Event: &rpcproto.Event_PeerStateChange{
						PeerStateChange: &rpcproto.PeerStateChangeEvent{
							NewState: p.State(),
						},
					},
				}

				if err := stream.Send(e); errors.Is(err, io.EOF) {
					continue
				} else if err != nil {
					s.logger.Error("Failed to send connection states", zap.Error(err))
				}
			}
		}

		return nil
	}); err != nil {
		s.logger.Error("Failed to send connection states", zap.Error(err))
	}
}

func (s *DaemonServer) OnPeerStateChanged(p *daemon.Peer, newState, prevState daemon.PeerState) {
	s.events.Send(&rpcproto.Event{
		Type: rpcproto.EventType_PEER_STATE_CHANGED,

		Interface: p.Interface.Name(),
		Peer:      p.PublicKey().Bytes(),

		Event: &rpcproto.Event_PeerStateChange{
			PeerStateChange: &rpcproto.PeerStateChangeEvent{
				NewState:  newState,
				PrevState: prevState,
			},
		},
	})
}

func serializeToString(in any) (string, error) {
	if tm, ok := in.(encoding.TextMarshaler); ok {
		b, err := tm.MarshalText()
		if err != nil {
			return "", err
		}

		return string(b), nil
	}

	return fmt.Sprint(in), nil
}

func settingToValue(val any) (*rpcproto.ConfigValue, error) {
	cval := &rpcproto.ConfigValue{}

	if val == nil {
		return cval, nil
	} else if rval := reflect.ValueOf(val); rval.Kind() == reflect.Slice {
		for i := range rval.Len() {
			e := rval.Index(i)

			s, err := serializeToString(e.Interface())
			if err != nil {
				return nil, err
			}

			cval.List = append(cval.List, s)
		}
	} else {
		cval.Scalar = fmt.Sprint(val)
	}

	return cval, nil
}

func decodeError(err error) error {
	var msErr *mapstructure.Error

	if errors.As(err, &msErr) {
		sts := status.New(codes.InvalidArgument, "Failed to decode")

		for _, err := range msErr.Errors {
			sts, _ = sts.WithDetails(&proto.Error{
				Message: err,
			})
		}

		return sts.Err()
	}

	return status.Error(codes.InvalidArgument, err.Error())
}
