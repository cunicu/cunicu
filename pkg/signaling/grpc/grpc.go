// Package grpc implements a signaling backend using a central gRPC service
package grpc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/stv0g/cunicu/pkg/util/buildinfo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	credsinsecure "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func ParseURL(urlStr string) (string, []grpc.DialOption, error) {
	opts := []grpc.DialOption{}

	u, err := url.Parse(urlStr)
	if err != nil {
		return "", nil, err
	}

	q := u.Query()

	insecure := false
	if q.Has("insecure") {
		var err error
		if insecure, err = strconv.ParseBool(q.Get("insecure")); err != nil {
			return "", nil, fmt.Errorf("failed to parse 'insecure' option: %w", err)
		}
	}

	skipVerify := false
	if q.Has("skip_verify") {
		var err error
		if skipVerify, err = strconv.ParseBool(q.Get("skip_verify")); err != nil {
			return "", nil, fmt.Errorf("failed to parse 'skip_verify' option: %w", err)
		}
	}

	var creds credentials.TransportCredentials
	if insecure {
		creds = credsinsecure.NewCredentials()
	} else {
		// Use system certificate store
		cfg := &tls.Config{
			//#nosec G402 -- Users should have the freedom to disable verification for self-signed certificates
			InsecureSkipVerify: skipVerify,
		}

		if fn := os.Getenv("SSLKEYLOGFILE"); fn != "" {
			var err error

			//#nosec G304 -- Filename is only controlled by env var
			if cfg.KeyLogWriter, err = os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600); err != nil {
				return "", nil, fmt.Errorf("failed to open SSL keylog file: %w", err)
			}
		}

		creds = credentials.NewTLS(cfg)
	}

	opts = append(opts,
		grpc.WithTransportCredentials(creds),
		grpc.WithUserAgent(buildinfo.UserAgent()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time: 10 * time.Second,
		}),
	)

	if u.Host == "" {
		return "", nil, errors.New("missing gRPC server url")
	}

	return u.Host, opts, nil
}
