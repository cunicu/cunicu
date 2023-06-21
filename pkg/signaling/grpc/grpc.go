// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package grpc implements a signaling backend using a central gRPC service
package grpc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	credsinsecure "google.golang.org/grpc/credentials/insecure"

	"github.com/stv0g/cunicu/pkg/buildinfo"
)

var errInvalidServerHostname = errors.New("missing gRPC server url")

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
			// Users should have the freedom to disable verification for self-signed certificates
			//nolint:gosec
			InsecureSkipVerify: skipVerify,
		}

		if fn := os.Getenv("SSLKEYLOGFILE"); fn != "" {
			var err error

			if cfg.KeyLogWriter, err = os.OpenFile(fn, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600); err != nil {
				return "", nil, fmt.Errorf("failed to open SSL keylog file: %w", err)
			}
		}

		creds = credentials.NewTLS(cfg)
	}

	opts = append(opts,
		grpc.WithTransportCredentials(creds),
		grpc.WithUserAgent(buildinfo.UserAgent()),
	)

	if u.Host == "" {
		return "", nil, errInvalidServerHostname
	}

	return u.Host, opts, nil
}
