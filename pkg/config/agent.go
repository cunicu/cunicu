// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/pion/ice/v2"
	"github.com/pion/stun"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/pkg/log"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
	grpcx "github.com/stv0g/cunicu/pkg/signaling/grpc"
	"github.com/stv0g/cunicu/pkg/types/slices"
)

var errInvalidURLScheme = errors.New("invalid ICE URL scheme")

func (c *InterfaceSettings) AgentURLs(ctx context.Context, pk *crypto.Key) ([]*stun.URI, error) { //nolint:gocognit
	logger := log.Global.Named("config")

	iceURLs := []*stun.URI{}

	g := errgroup.Group{}

	for _, u := range c.ICE.URLs {
		switch u.Scheme {
		case "stun", "stuns", "turn", "turns":
			// Extract credentials from URL
			// Warning: This is not standardized in RFCs 7064/7065
			iu, user, pass, _, err := icex.ParseURL(u.String())
			if err != nil {
				return nil, fmt.Errorf("failed to parse STUN/TURN URL '%s': %w", u.String(), err)
			}

			if user != "" && pass != "" {
				iu.Username = user
				iu.Password = pass
			} else {
				iu.Username = c.ICE.Username
				iu.Password = c.ICE.Password
			}

			iceURLs = append(iceURLs, iu)

		case "grpc":
			u := u
			g.Go(func() error {
				name, opts, err := grpcx.ParseURL(u.String())
				if err != nil {
					return err
				}

				conn, err := grpc.DialContext(ctx, name, opts...)
				if err != nil {
					return fmt.Errorf("failed to connect to gRPC server: %w", err)
				}

				defer conn.Close()

				client := signalingproto.NewRelayRegistryClient(conn)

				resp, err := client.GetRelays(ctx, &signalingproto.GetRelaysParams{
					PublicKey: pk.Bytes(),
				})
				if err != nil {
					return fmt.Errorf("received error from gRPC server: %w", err)
				}

				for _, svr := range resp.Relays {
					u, err := stun.ParseURI(svr.Url)
					if err != nil {
						return fmt.Errorf("failed to parse STUN/TURN URL '%s': %w", u, err)
					}

					u.Username = svr.Username
					u.Password = svr.Password

					iceURLs = append(iceURLs, u)
				}

				logger.Debug("Retrieved ICE servers from relay", zap.Any("uris", iceURLs))

				return nil
			})

		default:
			return nil, fmt.Errorf("%w: %s", errInvalidURLScheme, u.Scheme)
		}
	}

	return iceURLs, g.Wait()
}

func (c *InterfaceSettings) AgentConfig(ctx context.Context, peer *crypto.Key) (*ice.AgentConfig, error) { //nolint:gocognit
	var err error

	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.ICE.InsecureSkipVerify,
		Lite:               c.ICE.Lite,
		PortMin:            uint16(c.ICE.PortRange.Min),
		PortMax:            uint16(c.ICE.PortRange.Max),
	}

	cfg.InterfaceFilter = func(name string) bool {
		match, err := filepath.Match(c.ICE.InterfaceFilter, name)
		return err == nil && match
	}

	// ICE URLs
	if len(c.ICE.URLs) > 0 && len(c.ICE.CandidateTypes) > 0 {
		needsURLs := false
		for _, ct := range c.ICE.CandidateTypes {
			if ct.CandidateType == ice.CandidateTypeRelay || ct.CandidateType == ice.CandidateTypeServerReflexive {
				needsURLs = true
			}
		}

		if needsURLs {
			if cfg.Urls, err = c.AgentURLs(ctx, peer); err != nil {
				return nil, fmt.Errorf("failed to gather ICE URLs: %w", err)
			}

			// Filter URLs
			cfg.Urls = slices.Filter(cfg.Urls, func(u *stun.URI) bool {
				if isRelay := u.Scheme == stun.SchemeTypeTURN || u.Scheme == stun.SchemeTypeTURNS; isRelay {
					if c.ICE.RelayTCP != nil && *c.ICE.RelayTCP && u.Proto == stun.ProtoTypeUDP {
						return false
					}

					if c.ICE.RelayTLS != nil && *c.ICE.RelayTLS && u.Scheme == stun.SchemeTypeTURN {
						return false
					}
				}

				return true
			})
		}
	}

	if len(c.ICE.NAT1to1IPs) > 0 {
		cfg.NAT1To1IPs = c.ICE.NAT1to1IPs
	}

	if mbr := uint16(c.ICE.MaxBindingRequests); mbr > 0 {
		cfg.MaxBindingRequests = &mbr
	}

	if c.ICE.MDNS {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if to := c.ICE.DisconnectedTimeout; to > 0 {
		cfg.DisconnectedTimeout = &to
	}

	if to := c.ICE.FailedTimeout; to > 0 {
		cfg.FailedTimeout = &to
	}

	if to := c.ICE.KeepaliveInterval; to > 0 {
		cfg.KeepaliveInterval = &to
	}

	if to := c.ICE.CheckInterval; to > 0 {
		cfg.CheckInterval = &to
	}

	if len(c.ICE.CandidateTypes) > 0 {
		cfg.CandidateTypes = []ice.CandidateType{}
		for _, t := range c.ICE.CandidateTypes {
			cfg.CandidateTypes = append(cfg.CandidateTypes, t.CandidateType)
		}
	}

	cfg.NetworkTypes = []ice.NetworkType{}
	for _, t := range c.ICE.NetworkTypes {
		cfg.NetworkTypes = append(cfg.NetworkTypes, t.NetworkType)
	}

	return cfg, nil
}
