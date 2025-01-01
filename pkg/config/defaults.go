// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pion/ice/v4"

	"cunicu.li/cunicu/pkg/log"
	"cunicu.li/cunicu/pkg/wg"
)

const (
	// Ephemeral Port Range (RFC6056 Sect. 2.1).
	EphemeralPortMin = (1 << 15) + (1 << 14)
	EphemeralPortMax = (1 << 16) - 1
)

//nolint:gochecknoglobals
var (
	DefaultPrefixes = []string{"fc2f:9a4d::/32", "10.237.0.0/16"}

	DefaultBackends = []url.URL{
		{
			Scheme: "grpc",
			Host:   "signal.cunicu.li:443",
		},
	}

	DefaultICEURLs = []url.URL{
		{
			Scheme: "grpc",
			Host:   "relay.cunicu.li:443",
		},
	}

	DefaultSettings = Settings{
		Backends: DefaultBackends,
		RPC: RPCSettings{
			Socket: DefaultSocketPath,
			Wait:   false,
		},
		Log: LogSettings{
			Banner: true,
		},
		WatchInterval: 1 * time.Second,
		DefaultInterfaceSettings: InterfaceSettings{
			DiscoverPeers:     true,
			DiscoverEndpoints: true,
			SyncConfig:        true,
			SyncHosts:         true,
			SyncRoutes:        true,
			WatchRoutes:       true,

			PortForwarding: true,

			ICE: ICESettings{
				URLs:                DefaultICEURLs,
				CheckInterval:       200 * time.Millisecond,
				DisconnectedTimeout: 5 * time.Second,
				FailedTimeout:       25 * time.Second,
				RestartTimeout:      10 * time.Second,
				InterfaceFilter:     "*",
				KeepaliveInterval:   2 * time.Second,
				MaxBindingRequests:  7,
				PortRange: PortRangeSettings{
					Min: EphemeralPortMin,
					Max: EphemeralPortMax,
				},
				CandidateTypes: []ice.CandidateType{
					ice.CandidateTypeHost,
					ice.CandidateTypeServerReflexive,
					ice.CandidateTypePeerReflexive,
					ice.CandidateTypeRelay,
				},
				NetworkTypes: []ice.NetworkType{
					ice.NetworkTypeUDP4,
					ice.NetworkTypeUDP6,
					ice.NetworkTypeTCP4,
					ice.NetworkTypeTCP6,
				},
			},

			RoutingTable: DefaultRouteTable,

			ListenPortRange: &PortRangeSettings{
				Min: wg.DefaultPort,
				Max: EphemeralPortMax,
			},
		},
	}
)

func InitDefaults() error {
	logger := log.Global.Named("config")

	s := &DefaultSettings.DefaultInterfaceSettings

	// Check if WireGuard interface can be created by the kernel
	if !s.UserSpace && !wg.KernelModuleExists() {
		logger.Warn("The system does not have kernel support for WireGuard. Falling back to user-space implementation.")

		s.UserSpace = true
	}

	// Set default hostname
	if s.HostName == "" {
		hn, err := os.Hostname()
		if err != nil {
			return fmt.Errorf("failed to get hostname: %w", err)
		}

		// Do not use FQDN, but just the hostname part
		s.HostName = strings.Split(hn, ".")[0]
	}

	for _, pfxStr := range DefaultPrefixes {
		_, pfx, _ := net.ParseCIDR(pfxStr)
		s.Prefixes = append(s.Prefixes, *pfx)
	}

	return nil
}
