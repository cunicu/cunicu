// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/pion/ice/v2"

	"github.com/stv0g/cunicu/pkg/crypto"
	icex "github.com/stv0g/cunicu/pkg/ice"
)

var errInvalidSettings = errors.New("invalid settings")

//nolint:revive
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

	RelayTCP *bool `koanf:"relay_tcp,omitempty"`
	RelayTLS *bool `koanf:"relay_tls,omitempty"`

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

func (s *ICESettings) HasCandidateType(ct ice.CandidateType) bool {
	for _, c := range s.CandidateTypes {
		if ct == c.CandidateType {
			return true
		}
	}

	return false
}

func (s *ICESettings) HasNetworkType(nt ice.NetworkType) bool {
	for _, n := range s.NetworkTypes {
		if nt == n.NetworkType {
			return true
		}
	}

	return false
}

type RPCSettings struct {
	Socket string `koanf:"socket,omitempty"`
	Wait   bool   `koanf:"wait,omitempty"`
}

type HookSetting any

type PeerSettings struct {
	PublicKey                   crypto.Key           `koanf:"public_key,omitempty"`
	PresharedKey                crypto.Key           `koanf:"preshared_key,omitempty"`
	PresharedKeyPassphrase      crypto.KeyPassphrase `koanf:"preshared_key_passphrase,omitempty"`
	Endpoint                    string               `koanf:"endpoint,omitempty"`
	PersistentKeepaliveInterval time.Duration        `koanf:"persistent_keepalive,omitempty"`
	AllowedIPs                  []net.IPNet          `koanf:"allowed_ips,omitempty"`
}

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
	HostName string `koanf:"hostname,omitempty"`
	Domain   string `koanf:"domain,omitempty"`

	ExtraHosts map[string][]net.IPAddr `koanf:"extra_hosts,omitempty"`

	MTU       int          `koanf:"mtu,omitempty"`
	DNS       []net.IPAddr `koanf:"dns,omitempty"`
	Addresses []net.IPNet  `koanf:"addresses,omitempty"`
	Prefixes  []net.IPNet  `koanf:"prefixes"`
	Networks  []net.IPNet  `koanf:"networks,omitempty"`

	// Peer discovery
	Community crypto.KeyPassphrase `koanf:"community,omitempty"`
	Whitelist []crypto.Key         `koanf:"whitelist,omitempty"`
	Blacklist []crypto.Key         `koanf:"blacklist,omitempty"`

	// Endpoint discovery
	ICE            ICESettings `koanf:"ice,omitempty"`
	PortForwarding bool        `koanf:"port_forwarding,omitempty"`

	// Route sync
	RoutingTable int `koanf:"routing_table,omitempty"`

	// Hooks
	Hooks []HookSetting `koanf:"hooks,omitempty"`

	// WireGuard
	UserSpace       bool                    `koanf:"userspace,omitempty"`
	PrivateKey      crypto.Key              `koanf:"private_key,omitempty"`
	ListenPort      *int                    `koanf:"listen_port,omitempty"`
	ListenPortRange *PortRangeSettings      `koanf:"listen_port_range,omitempty"`
	FirewallMark    int                     `koanf:"fwmark,omitempty"`
	Peers           map[string]PeerSettings `koanf:"peers,omitempty"`

	// Feature flags
	DiscoverEndpoints bool `koanf:"discover_endpoints,omitempty"`
	DiscoverPeers     bool `koanf:"discover_peers,omitempty"`
	SyncConfig        bool `koanf:"sync_config,omitempty"`
	SyncRoutes        bool `koanf:"sync_routes,omitempty"`
	SyncHosts         bool `koanf:"sync_hosts,omitempty"`

	WatchConfig bool `koanf:"watch_config,omitempty"`
	WatchRoutes bool `koanf:"watch_routes,omitempty"`
}

type Settings struct {
	Experimental bool `koanf:"experimental,omitempty"`

	WatchInterval time.Duration `koanf:"watch_interval,omitempty"`
	Backends      []BackendURL  `koanf:"backends,omitempty"`

	RPC    RPCSettings    `koanf:"rpc,omitempty"`
	Config ConfigSettings `koanf:"config,omitempty"`

	DefaultInterfaceSettings InterfaceSettings            `koanf:",squash"`
	Interfaces               map[string]InterfaceSettings `koanf:"interfaces"`
}

// Check performs plausibility checks on the provided configuration.
func (c *Settings) Check() error {
	if err := c.DefaultInterfaceSettings.Check(); err != nil {
		return err
	}

	for _, icfg := range c.Interfaces {
		if err := icfg.Check(); err != nil {
			return err
		}
	}

	return nil
}

func (c *InterfaceSettings) Check() error {
	if c.ListenPortRange != nil && c.ListenPortRange.Min > c.ListenPortRange.Max {
		return fmt.Errorf("%w: WireGuard minimal listen port (%d) must be smaller or equal than maximal port (%d)",
			errInvalidSettings,
			c.ListenPortRange.Min,
			c.ListenPortRange.Max,
		)
	}

	return nil
}
