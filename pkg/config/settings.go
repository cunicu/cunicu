package config

import (
	"net"
	"path/filepath"
	"time"

	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
)

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
	Enabled bool   `koanf:"enabled,omitempty"`
	Path    string `koanf:"path,omitempty"`
	Watch   bool   `koanf:"watch,omitempty"`
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
	UserSpace       bool                    `koanf:"userspace,omitempty"`
	PrivateKey      crypto.Key              `koanf:"private_key,omitempty"`
	ListenPort      *int                    `koanf:"listen_port,omitempty"`
	ListenPortRange *PortRangeSettings      `koanf:"listen_port_range,omitempty"`
	FirewallMark    int                     `koanf:"fwmark,omitempty"`
	DNS             []net.IP                `koanf:"dns,omitempty"`
	Peers           []WireGuardPeerSettings `koanf:"peers,omitempty"`
}

type AutoConfigSettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	MTU                int         `koanf:"mtu,omitempty"`
	Addresses          []net.IPNet `koanf:"addresses,omitempty"`
	LinkLocalAddresses bool        `koanf:"link_local,omitempty"`
}

type HostSyncSettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	Domain string `koanf:"domain,omitempty"`
}

type PeerDiscoverySettings struct {
	Enabled bool `koanf:"enabled,omitempty"`

	Hostname  string               `koanf:"hostname,omitempty"`
	Community crypto.KeyPassphrase `koanf:"community,omitempty"`
	Whitelist []crypto.Key         `koanf:"whitelist,omitempty"`
	Networks  []net.IPNet          `koanf:"networks,omitempty"`
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
	Name    string
	Pattern string

	WireGuard    WireGuardSettings         `koanf:"wireguard,omitempty"`
	AutoConfig   AutoConfigSettings        `koanf:"autocfg,omitempty"`
	ConfigSync   ConfigSyncSettings        `koanf:"cfgsync,omitempty"`
	RouteSync    RouteSyncSettings         `koanf:"rtsync,omitempty"`
	HostSync     HostSyncSettings          `koanf:"hsync,omitempty"`
	EndpointDisc EndpointDiscoverySettings `koanf:"epdisc,omitempty"`
	PeerDisc     PeerDiscoverySettings     `koanf:"pdisc,omitempty"`
}

func (i *InterfaceSettings) Matches(name string) bool {
	if i.Pattern != "" {
		if ok, err := filepath.Match(i.Pattern, name); err == nil {
			return ok
		}
	} else if i.Name != "" {
		return name == i.Name
	}

	return false
}

type Settings struct {
	WatchInterval time.Duration `koanf:"watch_interval,omitempty"`
	RPC           RPCSettings   `koanf:"rpc,omitempty"`
	Backends      []BackendURL  `koanf:"backends,omitempty"`

	// Hooks are a global setting and not currently not customizable per interface.
	Hooks []HookSetting `koanf:"hooks,omitempty"`

	DefaultInterfaceSettings InterfaceSettings            `koanf:",squash"`
	Interfaces               map[string]InterfaceSettings `koanf:"interfaces"`
}
