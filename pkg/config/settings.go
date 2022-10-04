package config

import (
	"fmt"
	"net"
	"time"

	"github.com/pion/ice/v2"
	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
	"golang.org/x/exp/maps"
)

type ConfigSettings struct {
	Watch bool `koanf:"watch,omitempty"`
}

type PortRangeSettings struct {
	Min int `koanf:"min,omitempty"`
	Max int `koanf:"max,omitempty"`
}

type ICESettings struct {
	URLs           []URL                `koanf:"urls,omitempty"`
	CandidateTypes []icex.CandidateType `koanf:"candidate_types,omitempty"`
	NetworkTypes   []icex.NetworkType   `koanf:"network_types,omitempty"`
	NAT1to1IPs     []string             `koanf:"nat_1to1_ips,omitempty"`

	PortRange PortRangeSettings `koanf:"port_range,omitempty"`

	Lite               bool `koanf:"lite,omitempty"`
	MDNS               bool `koanf:"mdns,omitempty"`
	MaxBindingRequests int  `koanf:"max_binding_requests,omitempty"`
	InsecureSkipVerify bool `koanf:"insecure_skip_verify,omitempty"`

	InterfaceFilter string `koanf:"interface_filter,omitempty"`

	DisconnectedTimeout time.Duration `koanf:"disconnected_timeout,omitempty"`
	FailedTimeout       time.Duration `koanf:"failed_timeout,omitempty"`

	// KeepaliveInterval used to keep candidates alive
	KeepaliveInterval time.Duration `koanf:"keepalive_interval,omitempty"`

	// CheckInterval is the interval at which the agent performs candidate checks in the connecting phase
	CheckInterval  time.Duration `koanf:"check_interval,omitempty"`
	RestartTimeout time.Duration `koanf:"restart_timeout,omitempty"`

	Username string `koanf:"username,omitempty"`
	Password string `koanf:"password,omitempty"`
}

type RPCSettings struct {
	Socket string `koanf:"socket,omitempty"`
	Wait   bool   `koanf:"wait,omitempty"`
}

type ConfigSyncSettings struct {
	Enabled bool `koanf:"enabled,omitempty"`
}

type RouteSyncSettings struct {
	Enabled bool `koanf:"enabled,omitempty"`
	Table   int  `koanf:"table,omitempty"`
	Watch   bool `koanf:"watch,omitempty"`
}

type WireGuardPeerSettings struct {
	PublicKey                   crypto.Key           `koanf:"public_key,omitempty"`
	PresharedKey                crypto.Key           `koanf:"preshared_key,omitempty"`
	PresharedKeyPassphrase      crypto.KeyPassphrase `koanf:"preshared_key_passphrase,omitempty"`
	Endpoint                    string               `koanf:"endpoint,omitempty"`
	PersistentKeepaliveInterval time.Duration        `koanf:"persistent_keepalive,omitempty"`
	AllowedIPs                  []net.IPNet          `koanf:"allowed_ips,omitempty"`
}

type WireGuardSettings struct {
	UserSpace       bool                             `koanf:"userspace,omitempty"`
	PrivateKey      crypto.Key                       `koanf:"private_key,omitempty"`
	ListenPort      *int                             `koanf:"listen_port,omitempty"`
	ListenPortRange *PortRangeSettings               `koanf:"listen_port_range,omitempty"`
	FirewallMark    int                              `koanf:"fwmark,omitempty"`
	Peers           map[string]WireGuardPeerSettings `koanf:"peers,omitempty"`
}

type AutoConfigSettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	DNS       []net.IPAddr `koanf:"dns,omitempty"`
	MTU       int          `koanf:"mtu,omitempty"`
	Addresses []net.IPNet  `koanf:"addresses,omitempty"`
	Prefixes  []net.IPNet  `koanf:"prefixes"`
}

type HostSyncSettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	Domain string `koanf:"domain,omitempty"`
}

type PeerDiscoverySettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	Name      string               `koanf:"hostname,omitempty"`
	Community crypto.KeyPassphrase `koanf:"community,omitempty"`
	Networks  []net.IPNet          `koanf:"networks,omitempty"`
	Whitelist []crypto.Key         `koanf:"whitelist,omitempty"`
	Blacklist []crypto.Key         `koanf:"blacklist,omitempty"`
}

type EndpointDiscoverySettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	ICE ICESettings `koanf:"ice,omitempty"`
}

type HookSetting any

type BaseHookSetting struct {
	Type string `koanf:"type"`
}

type WebHookSetting struct {
	BaseHookSetting `koanf:",squash"`
	URL             URL               `koanf:"url"`
	Method          string            `koanf:"method"`
	Headers         map[string]string `koanf:"headers"`
}

type ExecHookSetting struct {
	BaseHookSetting `koanf:",squash"`
	Command         string            `koanf:"command"`
	Args            []string          `koanf:"args"`
	Env             map[string]string `koanf:"env"`
	Stdin           bool              `koanf:"stdin"`
}

type InterfaceSettings struct {
	AutoConfig   AutoConfigSettings        `koanf:"autocfg,omitempty"`
	ConfigSync   ConfigSyncSettings        `koanf:"cfgsync,omitempty"`
	EndpointDisc EndpointDiscoverySettings `koanf:"epdisc,omitempty"`
	Hooks        []HookSetting             `koanf:"hooks,omitempty"`
	HostSync     HostSyncSettings          `koanf:"hsync,omitempty"`
	PeerDisc     PeerDiscoverySettings     `koanf:"pdisc,omitempty"`
	RouteSync    RouteSyncSettings         `koanf:"rtsync,omitempty"`
	WireGuard    WireGuardSettings         `koanf:"wireguard,omitempty"`
}

type Settings struct {
	WatchInterval time.Duration `koanf:"watch_interval,omitempty"`
	Backends      []BackendURL  `koanf:"backends,omitempty"`

	RPC    RPCSettings    `koanf:"rpc,omitempty"`
	Config ConfigSettings `koanf:"config,omitempty"`

	DefaultInterfaceSettings InterfaceSettings            `koanf:",squash"`
	Interfaces               map[string]InterfaceSettings `koanf:"interfaces"`
}

// Check performs plausibility checks on the provided configuration.
func (c *Settings) Check() error {
	icfgs := []InterfaceSettings{c.DefaultInterfaceSettings}
	icfgs = append(icfgs, maps.Values(c.Interfaces)...)

	for _, icfg := range icfgs {
		if len(icfg.EndpointDisc.ICE.URLs) > 0 && len(icfg.EndpointDisc.ICE.CandidateTypes) > 0 {
			needsURL := false
			for _, ct := range icfg.EndpointDisc.ICE.CandidateTypes {
				if ct.CandidateType == ice.CandidateTypeRelay || ct.CandidateType == ice.CandidateTypeServerReflexive {
					needsURL = true
				}
			}

			if !needsURL {
				icfg.EndpointDisc.ICE.URLs = nil
			}
		}

		if icfg.WireGuard.ListenPortRange != nil && icfg.WireGuard.ListenPortRange.Min > icfg.WireGuard.ListenPortRange.Max {
			return fmt.Errorf("invalid settings: WireGuard minimal listen port (%d) must be smaller or equal than maximal port (%d)",
				icfg.WireGuard.ListenPortRange.Min,
				icfg.WireGuard.ListenPortRange.Max,
			)
		}
	}

	return nil
}
