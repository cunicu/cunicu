package config

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pion/ice/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	signalingproto "github.com/stv0g/cunicu/pkg/proto/signaling"
	grpcx "github.com/stv0g/cunicu/pkg/signaling/grpc"
)

func (c *InterfaceSettings) AgentURLs(ctx context.Context, pk *crypto.Key) ([]*ice.URL, error) {
	iceURLs := []*ice.URL{}

	g := errgroup.Group{}

	for _, u := range c.EndpointDisc.ICE.URLs {
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
				iu.Username = c.EndpointDisc.ICE.Username
				iu.Password = c.EndpointDisc.ICE.Password
			}

			iceURLs = append(iceURLs, iu)

		case "grpc":
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
					u, err := ice.ParseURL(svr.Url)
					if err != nil {
						return fmt.Errorf("failed to parse STUN/TURN URL '%s': %w", u, err)
					}

					u.Username = svr.Username
					u.Password = svr.Password

					iceURLs = append(iceURLs, u)
				}

				return nil
			})

		default:
			return nil, fmt.Errorf("invalid ICE URL scheme: %s", u.Scheme)
		}
	}

	return iceURLs, g.Wait()
}

func (c *InterfaceSettings) AgentConfig(ctx context.Context, peer *crypto.Key) (*ice.AgentConfig, error) {
	var err error

	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.EndpointDisc.ICE.InsecureSkipVerify,
		Lite:               c.EndpointDisc.ICE.Lite,
		PortMin:            uint16(c.EndpointDisc.ICE.PortRange.Min),
		PortMax:            uint16(c.EndpointDisc.ICE.PortRange.Max),
	}

	cfg.InterfaceFilter = func(name string) bool {
		match, err := filepath.Match(c.EndpointDisc.ICE.InterfaceFilter, name)
		return err == nil && match
	}

	// ICE URLs
	if cfg.Urls, err = c.AgentURLs(ctx, peer); err != nil {
		return nil, fmt.Errorf("failed to gather ICE URLs: %w", err)
	}

	if len(c.EndpointDisc.ICE.NAT1to1IPs) > 0 {
		cfg.NAT1To1IPs = c.EndpointDisc.ICE.NAT1to1IPs
	}

	if mbr := uint16(c.EndpointDisc.ICE.MaxBindingRequests); mbr > 0 {
		cfg.MaxBindingRequests = &mbr
	}

	if c.EndpointDisc.ICE.MDNS {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if to := c.EndpointDisc.ICE.DisconnectedTimeout; to > 0 {
		cfg.DisconnectedTimeout = &to
	}

	if to := c.EndpointDisc.ICE.FailedTimeout; to > 0 {
		cfg.FailedTimeout = &to
	}

	if to := c.EndpointDisc.ICE.KeepaliveInterval; to > 0 {
		cfg.KeepaliveInterval = &to
	}

	if to := c.EndpointDisc.ICE.CheckInterval; to > 0 {
		cfg.CheckInterval = &to
	}

	if len(c.EndpointDisc.ICE.CandidateTypes) > 0 {
		cfg.CandidateTypes = []ice.CandidateType{}
		for _, t := range c.EndpointDisc.ICE.CandidateTypes {
			cfg.CandidateTypes = append(cfg.CandidateTypes, t.CandidateType)
		}
	}

	if len(c.EndpointDisc.ICE.NetworkTypes) > 0 {
		cfg.NetworkTypes = []ice.NetworkType{}
		for _, t := range c.EndpointDisc.ICE.NetworkTypes {
			cfg.NetworkTypes = append(cfg.NetworkTypes, t.NetworkType)
		}
	} else {
		cfg.NetworkTypes = []ice.NetworkType{
			ice.NetworkTypeTCP4,
			ice.NetworkTypeUDP4,
			ice.NetworkTypeTCP6,
			ice.NetworkTypeUDP6,
		}
	}

	return cfg, nil
}
