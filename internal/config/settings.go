package config

import (
	"io"
	"time"

	"gopkg.in/yaml.v3"

	icex "riasc.eu/wice/internal/ice"
)

type PortSettings struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}

type ICESettings struct {
	URLs           []icex.URL           `yaml:"urls,omitempty"`
	CandidateTypes []icex.CandidateType `yaml:"candidate_types,omitempty"`
	NetworkTypes   []icex.NetworkType   `yaml:"network_types,omitempty"`
	NAT1to1IPs     []string             `yaml:"nat_1to1_ips,omitempty"`

	Port PortSettings `yaml:"port,omitempty"`

	Lite               bool `yaml:"lite,omitempty"`
	MDNS               bool `yaml:"mdns,omitempty"`
	MaxBindingRequests int  `yaml:"max_binding_requests,omitempty"`
	InsecureSkipVerify bool `yaml:"insecure_skip_verify,omitempty"`

	InterfaceFilter Regexp `yaml:"interface_filter,omitempty"`

	DisconnectedTimeout time.Duration `yaml:"disconnected_timeout,omitempty"`
	FailedTimeout       time.Duration `yaml:"failed_timeout,omitempty"`

	// KeepaliveInterval used to keep candidates alive
	KeepaliveInterval time.Duration `yaml:"keepalive_interval,omitempty"`

	// CheckInterval is the interval at which the agent performs candidate checks in the connecting phase
	CheckInterval  time.Duration `yaml:"check_interval,omitempty"`
	RestartTimeout time.Duration `yaml:"restart_timeout,omitempty"`

	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
}

type ProxySettings struct {
	NFT  bool `yaml:"nft"`
	EBPF bool `yaml:"ebpf"`
}

type SocketSettings struct {
	Path string `yaml:"path,omitempty"`
	Wait bool   `yaml:"wait,omitempty"`
}

type WireguardConfigSettings struct {
	Path string `yaml:"path,omitempty"`
	Sync bool   `yaml:"sync,omitempty"`
}

type WireguardPortSettings struct {
	Min int `yaml:"min,omitempty"`
	Max int `yaml:"max,omitempty"`
}

type WireguardSettings struct {
	Config WireguardConfigSettings `yaml:"config,omitempty"`

	Port PortSettings `yaml:"port,omitempty"`

	Userspace       bool     `yaml:"userspace,omitempty"`
	InterfaceFilter Regexp   `yaml:"interface_filter,omitempty"`
	Interfaces      []string `yaml:"interfaces,omitempty"`
}

type Settings struct {
	Community     string        `yaml:"community,omitempty"`
	WatchInterval time.Duration `yaml:"watch_interval,omitempty"`

	Backends []BackendURL `yaml:"backends,omitempty"`

	ICE       ICESettings       `yaml:"ice,omitempty"`
	Proxy     ProxySettings     `yaml:"proxy,omitempty"`
	Socket    SocketSettings    `yaml:"socket,omitempty"`
	Wireguard WireguardSettings `yaml:"wg,omitempty"`
}

func (s *Settings) Dump(wr io.Writer) error {
	enc := yaml.NewEncoder(wr)
	enc.SetIndent(2)

	return enc.Encode(s)
}
