// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec
	"encoding/base64"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/proto"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
)

const (
	DefaultRelayTTL = 1 * time.Hour
)

type RelayInfo struct {
	URL   string
	Realm string

	Username string
	Password string

	TTL    time.Duration
	Secret string
}

func NewRelayInfo(arg string) (RelayInfo, error) {
	u, user, pass, q, err := icex.ParseURL(arg)
	if err != nil {
		return RelayInfo{}, fmt.Errorf("invalid URL: %w", err)
	}

	r := RelayInfo{
		URL:      u.String(),
		Secret:   q.Get("secret"),
		Username: user,
		Password: pass,
		TTL:      DefaultRelayTTL,
	}

	if t := q.Get("ttl"); t != "" {
		ttl, err := time.ParseDuration(t)
		if err != nil {
			return RelayInfo{}, fmt.Errorf("invalid TTL: %w", err)
		}

		r.TTL = ttl
	}

	return r, nil
}

func NewRelayInfos(args []string) ([]RelayInfo, error) {
	relays := []RelayInfo{}
	for _, arg := range args {
		relay, err := NewRelayInfo(arg)
		if err != nil {
			return nil, err
		}

		relays = append(relays, relay)
	}

	return relays, nil
}

func (s *RelayInfo) GetCredentials(username string) (string, string, time.Time) {
	if s.Username != "" && s.Password != "" {
		return s.Username, s.Password, time.Time{}
	} else if s.Secret != "" {
		if s.Username != "" {
			username = s.Username
		}

		exp := time.Now().Add(s.TTL)
		user := fmt.Sprintf("%d:%s", exp.Unix(), username)

		digest := hmac.New(sha1.New, []byte(s.Secret))
		digest.Write([]byte(user))

		passRaw := digest.Sum(nil)
		pass := base64.StdEncoding.EncodeToString(passRaw)

		return user, pass, exp
	}

	return "", "", time.Time{}
}

type RelayAPIServer struct {
	signalingproto.UnimplementedRelayRegistryServer

	relays []RelayInfo

	*grpc.Server

	logger *log.Logger
}

func NewRelayAPIServer(relays []RelayInfo, opts ...grpc.ServerOption) (*RelayAPIServer, error) {
	s := &RelayAPIServer{
		relays: relays,
		Server: grpc.NewServer(opts...),
		logger: log.Global.Named("grpc.relay"),
	}

	signalingproto.RegisterRelayRegistryServer(s, s)

	return s, nil
}

func (s *RelayAPIServer) GetRelays(_ context.Context, params *signalingproto.GetRelaysParams) (*signalingproto.GetRelaysResp, error) {
	resp := &signalingproto.GetRelaysResp{}

	pk, err := crypto.ParseKeyBytes(params.PublicKey)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse public key: %s", err)
	}

	for _, svr := range s.relays {
		s := &signalingproto.RelayInfo{
			Url: svr.URL,
		}

		if svr.Secret != "" {
			var exp time.Time
			s.Username, s.Password, exp = svr.GetCredentials(pk.String())
			s.Expires = proto.Time(exp)
		}

		resp.Relays = append(resp.Relays, s)
	}

	return resp, nil
}

func (s *RelayAPIServer) Close() error {
	s.Server.GracefulStop()

	return nil
}
