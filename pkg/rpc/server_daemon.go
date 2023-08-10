// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
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

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/log"
	osx "github.com/stv0g/cunicu/pkg/os"
	"github.com/stv0g/cunicu/pkg/proto"
	coreproto "github.com/stv0g/cunicu/pkg/proto/core"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
	slicesx "github.com/stv0g/cunicu/pkg/types/slices"
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
	var err error
	var pk crypto.Key

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
	errs := []error{}
	settings := map[string]any{}

	numChanges := 0

	for key, value := range p.Settings {
		switch key {
		case "log.rules":
			rules := strings.Split(value, ",")
			filter, err := log.ParseFilter(rules)
			if err != nil {
				errs = append(errs, err)
			}
			log.UpdateFilter(filter)
			numChanges++

		default:
			if value == "" { // Unset value
				settings[key] = nil
			} else {
				settings[key] = value
			}
		}
	}

	changes, err := s.Config.Update(settings)
	if err != nil {
		errs = append(errs, err)
	}

	numChanges += len(changes)
	if numChanges == 0 {
		errs = append(errs, errNoSettingChanged)
	}

	if len(errs) > 0 {
		errstrs := []string{}
		for _, err := range errs {
			errstrs = append(errstrs, err.Error())
		}

		return nil, status.Error(codes.InvalidArgument, strings.Join(errstrs, ", "))
	}

	if err := s.Config.SaveRuntime(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Failed to save runtime configuration: %s", err)
	}

	return &proto.Empty{}, nil
}

func (s *DaemonServer) GetConfig(_ context.Context, p *rpcproto.GetConfigParams) (*rpcproto.GetConfigResp, error) {
	settings := map[string]string{}

	match := func(key string) bool {
		return p.KeyFilter == "" || strings.HasPrefix(key, p.KeyFilter)
	}

	for key, value := range s.Config.All() {
		if match(key) {
			str, err := settingToString(value)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to marshal: %s", err)
			}

			settings[key] = str
		}
	}

	if match("log.rules") {
		settings["log.rules"] = log.CurrentFilter().String()
	}

	return &rpcproto.GetConfigResp{
		Settings: settings,
	}, nil
}

func (s *DaemonServer) GetCompletion(_ context.Context, params *rpcproto.GetCompletionParams) (*rpcproto.GetCompletionResp, error) {
	var options []string
	var flags cobra.ShellCompDirective

	if len(params.Cmd) < 2 || params.Cmd[0] != "cunicu" {
		flags = cobra.ShellCompDirectiveError
	} else if params.Cmd[1] == "config" {
		options, flags = s.getConfigCompletion(params.Cmd[2], params.Args, params.ToComplete)
	} else {
		flags = cobra.ShellCompDirectiveNoFileComp
	}

	return &rpcproto.GetCompletionResp{
		Options: options,
		Flags:   int32(flags),
	}, nil
}

func (s *DaemonServer) getConfigCompletion(cmd string, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var options []string
	if isValueCompletion := len(args) > 0; isValueCompletion {
		if cmd != "set" {
			return nil, cobra.ShellCompDirectiveNoFileComp
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

	return options, cobra.ShellCompDirectiveNoFileComp
}

func (s *DaemonServer) ReloadConfig(_ context.Context, _ *proto.Empty) (*proto.Empty, error) {
	if _, err := s.Config.Reload(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to reload configuration: %s", err)
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

func settingToString(value any) (string, error) {
	v := reflect.ValueOf(value)
	switch {
	case v.Kind() == reflect.Slice:
		s := []string{}
		for i := 0; i < v.Len(); i++ {
			e := v.Index(i)
			in := e.Interface()

			if tm, ok := in.(encoding.TextMarshaler); ok {
				b, err := tm.MarshalText()
				if err != nil {
					return "", err
				}

				s = append(s, string(b))
			} else {
				s = append(s, fmt.Sprint(in))
			}
		}

		return strings.Join(s, ","), nil

	case value == nil:
		return "", nil

	default:
		return fmt.Sprintf("%v", value), nil
	}
}
