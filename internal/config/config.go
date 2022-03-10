package config

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	icex "riasc.eu/wice/internal/ice"

	"github.com/pion/ice/v2"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Copied from pion/ice/agent_config.go
const (
	// defaultCheckInterval is the interval at which the agent performs candidate checks in the connecting phase
	defaultCheckInterval = 200 * time.Millisecond

	// keepaliveInterval used to keep candidates alive
	defaultKeepaliveInterval = 2 * time.Second

	// defaultDisconnectedTimeout is the default time till an Agent transitions disconnected
	defaultDisconnectedTimeout = 5 * time.Second

	// defaultRestartInterval is the default time an Agent waits before it attempts an ICE restart
	defaultRestartTimeout = 5 * time.Second

	// defaultFailedTimeout is the default time till an Agent transitions to failed after disconnected
	defaultFailedTimeout = 5 * time.Second

	// max binding request before considering a pair failed
	defaultMaxBindingRequests = 7

	defaultWatchInterval = time.Second

	DefaultSocketPath = "/var/run/wice.sock"

	defaultWireguardConfigPath = "/etc/wireguard"
)

type Config struct {
	*viper.Viper

	flags  *pflag.FlagSet
	logger *zap.Logger

	WireguardInterfaces []string
	ConfigFiles         []string

	Backends []*url.URL
}

func Parse(args ...string) (*Config, error) {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	c := NewConfig(f)

	if err := c.flags.Parse(args); err != nil {
		return nil, err
	}

	if err := c.Setup(args); err != nil {
		return nil, err
	}

	return c, nil
}

func NewConfig(flags *pflag.FlagSet) *Config {
	c := &Config{
		Viper:       viper.New(),
		ConfigFiles: []string{},

		flags: flags,
	}

	c.SetDefault("ice.urls", []string{"stun:l.google.com:19302"})
	c.SetDefault("backends", []string{"p2p"})
	c.SetDefault("watch_interval", defaultWatchInterval)

	c.SetDefault("socket.path", DefaultSocketPath)

	c.SetDefault("ice.max_binding_requests", defaultMaxBindingRequests)

	c.SetDefault("ice.check_interval", defaultCheckInterval)
	c.SetDefault("ice.disconnected_timout", defaultDisconnectedTimeout)
	c.SetDefault("ice.failed_timeout", defaultFailedTimeout)
	c.SetDefault("ice.restart_timeout", defaultRestartTimeout)
	c.SetDefault("ice.keepalive_interval", defaultKeepaliveInterval)
	c.SetDefault("ice.nat_1to1_ips", []net.IP{})

	c.SetDefault("wg.config.path", defaultWireguardConfigPath)

	flags.StringP("config-domain", "A", "", "Perform auto-configuration via DNS")
	flags.StringSliceVarP(&c.ConfigFiles, "config", "c", []string{}, "Path of configuration files")

	flags.StringP("community", "x", "", "Community passphrase for discovering other peers")
	flags.StringSliceP("backend", "b", []string{}, "backend types / URLs")
	flags.StringP("interface-filter", "f", ".*", "regex for filtering Wireguard interfaces (e.g. \"wg-.*\")")
	flags.DurationP("watch-interval", "i", 0, "interval at which we are polling the kernel for updates on the Wireguard interfaces")

	flags.BoolP("wg-userspace", "u", false, "start userspace Wireguard daemon")
	flags.BoolP("wg-config-sync", "S", false, "sync Wireguard interface with configuration file (see \"wg synconf\")")
	flags.StringP("wg-config-path", "w", "", "base path to search for Wireguard configuration files")

	// ice.AgentConfig fields
	flags.StringSliceP("url", "a", []string{}, "STUN and/or TURN server addresses")
	flags.StringSlice("ice-candidate-type", []string{}, "usable candidate types (select from \"host\", \"srflx\", \"prflx\", \"relay\")")
	flags.StringSlice("ice-network-type", []string{}, "usable network types (select from \"udp4\", \"udp6\", \"tcp4\", \"tcp6\")")
	flags.StringSlice("ice-nat-1to1-ip", []string{}, "IP addresses which will be added as local server reflexive candidates")

	flags.Uint16("ice-port-min", 0, "minimum port for allocation policy (range: 0-65535)")
	flags.Uint16("ice-port-max", 0, "maximum port for allocation policy (range: 0-65535)")
	flags.BoolP("ice-lite", "L", false, "lite agents do not perform connectivity check and only provide host candidates")
	flags.BoolP("ice-mdns", "m", false, "enable local Multicast DNS discovery")
	flags.Uint16("ice-max-binding-requests", 0, "maximum number of binding request before considering a pair failed")
	flags.BoolP("ice-insecure-skip-verify", "k", false, "skip verification of TLS certificates for secure STUN/TURN servers")
	flags.String("ice-interface-filter", ".*", "regex for filtering local interfaces for ICE candidate gathering (e.g. \"eth[0-9]+\")")
	flags.Duration("ice-disconnected-timout", 0, "time till an Agent transitions disconnected")
	flags.Duration("ice-failed-timeout", 0, "time until an Agent transitions to failed after disconnected")
	flags.Duration("ice-keepalive-interval", 0, "interval netween STUN keepalives")
	flags.Duration("ice-check-interval", 0, "interval at which the agent performs candidate checks in the connecting phase")
	flags.Duration("ice-restart-timeout", 0, "time to wait before ICE restart")
	flags.StringP("ice-user", "U", "", "username for STUN/TURN credentials")
	flags.StringP("ice-pass", "P", "", "password for STUN/TURN credentials")

	flags.StringP("socket", "s", "", "Unix control and monitoring socket")
	flags.Bool("socket-wait", false, "wait until first client connected to control socket before continuing start")

	flagMap := map[string]string{
		"config-domain":            "domain",
		"wg-userspace":             "wg.userspace",
		"community":                "community",
		"backend":                  "backends",
		"interface-filter":         "wg.interface_filter",
		"watch-interval":           "watch_interval",
		"wg-config-sync":           "wg.config_sync",
		"wg-config-path":           "wg.config_path",
		"url":                      "ice.urls",
		"ice-candidate-type":       "ice.candidate_types",
		"ice-network-type":         "ice.network_types",
		"ice-nat-1to1-ip":          "ice.nat_1to1_ips",
		"ice-port-min":             "ice.port_min",
		"ice-port-max":             "ice.port_max",
		"ice-lite":                 "ice.lite",
		"ice-mdns":                 "ice.mdns",
		"ice-max-binding-requests": "ice.max_binding_requests",
		"ice-insecure-skip-verify": "ice.insecure_skip_verify",
		"ice-interface-filter":     "ice.interface_filter",
		"ice-disconnected-timout":  "ice.disconnected_timout",
		"ice-failed-timeout":       "ice.failed_timeout",
		"ice-keepalive-interval":   "ice.keepalive_interval",
		"ice-check-interval":       "ice.check_interval",
		"ice-restart-timeout":      "ice.restart_timeout",
		"ice-user":                 "ice.username",
		"ice-pass":                 "ice.password",
		"socket":                   "socket.path",
		"socket-wait":              "socket.wait",
	}

	flags.VisitAll(func(flag *pflag.Flag) {
		name := flag.Name
		if newName, ok := flagMap[name]; ok {
			c.BindPFlag(newName, flag)
		}
	})

	return c
}

func (c *Config) Setup(args []string) error {
	c.logger = zap.L().Named("config")

	c.WireguardInterfaces = args

	// First lookup settings via DNS
	if c.IsSet("domain") {
		domain := c.GetString("domain")
		if err := c.Lookup(domain); err != nil {
			return fmt.Errorf("DNS autoconfiguration failed: %w", err)
		}
	}

	if len(c.ConfigFiles) > 0 {
		// Merge config files from the flags.
		for _, file := range c.ConfigFiles {
			if u, err := url.Parse(file); err == nil && u.Scheme != "" {
				if err := c.MergeRemoteConfig(u); err != nil {
					return fmt.Errorf("failed to load remote config: %w", err)
				}

				c.logger.Debug("Using remote configuration file", zap.Any("url", u))
			} else {
				c.SetConfigFile(file)
				if err := c.MergeInConfig(); err != nil {
					return fmt.Errorf("failed to merge configurations: %w", err)
				}

				c.logger.Debug("Using configuration file", zap.String("file", c.ConfigFileUsed()))
			}
		}
	} else {
		c.AddConfigPath("/etc")
		c.AddConfigPath(filepath.Join("$HOME", ".config"))
		c.AddConfigPath(".")
		c.SetConfigName("wice")

		if err := c.MergeInConfig(); err == nil {
			c.logger.Debug("Using configuration file", zap.String("file", c.ConfigFileUsed()))
		} else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to merge configurations: %w", err)
		}
	}

	c.SetEnvPrefix("wice")
	c.AutomaticEnv()
	c.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := c.Load(); err != nil {
		return err
	}

	return nil
}

func (c *Config) MergeRemoteConfig(url *url.URL) error {
	if url.Scheme != "https" {
		host, _, _ := net.SplitHostPort(url.Host)
		ip, err := net.ResolveIPAddr("ip", host)
		if err != nil || !ip.IP.IsLoopback() {
			return errors.New("remote configuration must by provided via HTTPS")
		}
	}

	resp, err := http.DefaultClient.Get(url.String())
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", url, err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch: %s: %s", url, resp.Status)
	}

	return c.MergeConfig(resp.Body)
}

func (c *Config) Load() error {

	// Backends
	c.Backends = []*url.URL{}
	for _, u := range c.GetStringSlice("backends") {
		// Allow the user to specify just the backend type as a valid url.
		// E.g. "p2p" instead of "p2p:"
		if !strings.Contains(u, ":") {
			u += ":"
		}

		u, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("invalid backend URI: %w", err)
		}

		c.Backends = append(c.Backends, u)
	}

	return nil
}

func (c *Config) AgentConfig() (*ice.AgentConfig, error) {
	cfg := &ice.AgentConfig{
		InsecureSkipVerify: c.GetBool("ice.insecure_skip_verify"),
		Lite:               c.GetBool("ice.lite"),
		PortMin:            uint16(c.GetUint("ice.port.min")),
		PortMax:            uint16(c.GetUint("ice.port.max")),
	}

	interfaceFilterRegex, err := regexp.Compile(c.GetString("ice.interface_filter"))
	if err != nil {
		return nil, fmt.Errorf("invalid ice.interface_filter config: %w", err)
	}

	cfg.InterfaceFilter = func(name string) bool {
		return interfaceFilterRegex.Match([]byte(name))
	}

	// ICE URLS
	cfg.Urls = []*ice.URL{}
	for _, u := range c.GetStringSlice("ice.urls") {
		up, err := ice.ParseURL(u)
		if err != nil {
			return nil, fmt.Errorf("failed to parse ice.url: %s: %w", u, err)
		}

		cfg.Urls = append(cfg.Urls, up)
	}

	// Add default STUN/TURN servers
	// Set ICE credentials
	u := c.GetString("ice.username")
	p := c.GetString("ice.password")
	for _, q := range cfg.Urls {
		if u != "" {
			q.Username = u
		}

		if p != "" {
			q.Password = p
		}
	}

	if c.IsSet("ice.nat_1to1_ips") {
		cfg.NAT1To1IPs = c.GetStringSlice("ice.nat_1to1_ips")
	}

	if c.IsSet("ice.max_binding_requests") {
		i := uint16(c.GetInt("ice.max_binding_requests"))
		cfg.MaxBindingRequests = &i
	}

	if c.GetBool("ice.mdns") {
		cfg.MulticastDNSMode = ice.MulticastDNSModeQueryAndGather
	}

	if c.IsSet("ice.disconnected_timeout") {
		to := c.GetDuration("ice.disconnected_timeout")
		cfg.DisconnectedTimeout = &to
	}

	if c.IsSet("ice.failed_timeout") {
		to := c.GetDuration("ice.failed_timeout")
		cfg.FailedTimeout = &to
	}

	if c.IsSet("ice.keepalive_interval") {
		to := c.GetDuration("ice.keepalive_interval")
		cfg.KeepaliveInterval = &to
	}

	if c.IsSet("ice.check_interval") {
		to := c.GetDuration("ice.check_interval")
		cfg.CheckInterval = &to
	}

	// Filter candidate types
	candidateTypes := []ice.CandidateType{}
	for _, value := range c.GetStringSlice("ice.candidate_types") {
		ct, err := icex.CandidateTypeFromString(value)
		if err != nil {
			return nil, err
		}

		candidateTypes = append(candidateTypes, ct)
	}

	if len(candidateTypes) > 0 {
		cfg.CandidateTypes = candidateTypes
	}

	// Filter network types
	networkTypes := []ice.NetworkType{}
	for _, value := range c.GetStringSlice("ice.network_types") {
		ct, err := icex.NetworkTypeFromString(value)
		if err != nil {
			return nil, err
		}

		networkTypes = append(networkTypes, ct)
	}

	if len(networkTypes) > 0 {
		cfg.NetworkTypes = networkTypes
	}

	return cfg, nil
}

func (c *Config) Dump(wr io.Writer) error {
	return yaml.NewEncoder(wr).Encode(c.AllSettings())
}
