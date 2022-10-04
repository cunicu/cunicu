package grpc

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	logger *zap.Logger
}

func NewRelayAPIServer(relaysStrs []string, opts ...grpc.ServerOption) (*RelayAPIServer, error) {
	logger := zap.L().Named("grpc.server")

	relays, err := NewRelayInfos(relaysStrs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse relays: %w", err)
	}

	s := &RelayAPIServer{
		relays: relays,
		Server: grpc.NewServer(opts...),
		logger: logger,
	}

	signalingproto.RegisterRelayRegistryServer(s, s)

	return s, nil
}

func (s *RelayAPIServer) GetRelays(ctx context.Context, params *signalingproto.GetRelaysParams) (*signalingproto.GetRelaysResp, error) {
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
