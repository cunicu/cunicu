package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/pion/ice/v2"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

const (
	DefaultSocketPath = "/var/run/cunicu.sock"

	// Ephemeral Port Range (RFC6056 Sect. 2.1)
	EphemeralPortMin = (1 << 15) + (1 << 14)
	EphemeralPortMax = (1 << 16) - 1
)

var (
	DefaultPrefixes = []string{"fc2f:9a4d::/32", "10.237.0.0/16"}

	DefaultBackends = []BackendURL{
		{
			URL: url.URL{
				Scheme: "grpc",
				Host:   "signal.cunicu.li:443",
			},
		},
	}

	DefaultICEURLs = []URL{
		{url.URL{
			Scheme: "stun",
			Opaque: "stun.cunicu.li:3478",
		}},
		// TODO: Use relay API
		// {url.URL{
		// 	Scheme: "grpc",
		// 	Host:   "relay.cunicu.li:",
		// }},
	}

	DefaultSettings = Settings{
		Backends: DefaultBackends,
		RPC: RPCSettings{
			Socket: DefaultSocketPath,
			Wait:   false,
		},
		WatchInterval:            1 * time.Second,
		DefaultInterfaceSettings: DefaultInterfaceSettings,
	}

	DefaultInterfaceSettings = InterfaceSettings{
		AutoConfig: AutoConfigSettings{
			Enabled: true,
		},
		ConfigSync: ConfigSyncSettings{
			Enabled: true,
		},
		PeerDisc: PeerDiscoverySettings{
			Enabled: true,
		},
		EndpointDisc: EndpointDiscoverySettings{
			Enabled: true,

			ICE: ICESettings{
				URLs:                DefaultICEURLs,
				CheckInterval:       200 * time.Millisecond,
				DisconnectedTimeout: 5 * time.Second,
				FailedTimeout:       5 * time.Second,
				RestartTimeout:      5 * time.Second,
				InterfaceFilter:     "*",
				KeepaliveInterval:   2 * time.Second, // TODO: increase
				MaxBindingRequests:  7,
				PortRange: PortRangeSettings{
					Min: EphemeralPortMin,
					Max: EphemeralPortMax,
				},
				CandidateTypes: []icex.CandidateType{
					{CandidateType: ice.CandidateTypeHost},
					{CandidateType: ice.CandidateTypeServerReflexive},
					{CandidateType: ice.CandidateTypePeerReflexive},
					{CandidateType: ice.CandidateTypeRelay},
				},
				NetworkTypes: []icex.NetworkType{
					{NetworkType: ice.NetworkTypeUDP4},
					{NetworkType: ice.NetworkTypeUDP6},
					{NetworkType: ice.NetworkTypeTCP4},
					{NetworkType: ice.NetworkTypeTCP6},
				},
			},
		},
		HostSync: HostSyncSettings{
			Enabled: true,
		},
		RouteSync: RouteSyncSettings{
			Enabled: true,

			Watch: true,
			Table: DefaultRouteTable,
		},
		WireGuard: WireGuardSettings{
			ListenPortRange: &PortRangeSettings{
				Min: wg.DefaultPort,
				Max: EphemeralPortMax,
			},
		},
	}
)

func InitDefaults() error {
	var err error

	logger := zap.L().Named("config")

	s := &DefaultSettings.DefaultInterfaceSettings

	// Check if WireGuard interface can be created by the kernel
	if !s.WireGuard.UserSpace && !wg.KernelModuleExists() {
		logger.Warn("The system does not have kernel support for WireGuard. Falling back to user-space implementation.")
		s.WireGuard.UserSpace = true
	}

	// Set default hostname
	if s.PeerDisc.Name == "" {
		if s.PeerDisc.Name, err = os.Hostname(); err != nil {
			return fmt.Errorf("failed to get hostname: %w", err)
		}
	}

	for _, pfxStr := range DefaultPrefixes {
		_, pfx, _ := net.ParseCIDR(pfxStr)
		s.AutoConfig.Prefixes = append(s.AutoConfig.Prefixes, *pfx)
	}

	return nil
}
