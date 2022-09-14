// Package config defines, loads and parses project wide configuration settings from various sources
package config

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/imdario/mergo"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/mitchellh/mapstructure"
	"github.com/pion/ice/v2"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	envPrefix = "CUNICU_"

	// Map flags from the flags to to Koanf settings
	flagMap = map[string]string{
		// Config sync
		"cfgsync":       "cfgsync.enabled",
		"cfgsync-path":  "cfgsync.path",
		"cfgsync-watch": "cfgsync.watch",

		// Host sync
		"hsync": "hsync.enabled",

		// Route sync
		"rtsync":       "rtsync.enabled",
		"rtsync-table": "rtsync.table",

		"backend":        "backends",
		"watch-interval": "watch_interval",

		// Socket
		"rpc-socket": "rpc.socket",
		"rpc-wait":   "rpc.wait",

		// WireGuard
		"wg-userspace": "wireguard.userspace",

		// Endpoint discovery
		"epdisc":             "epdisc.enabled",
		"url":                "epdisc.ice.urls",
		"username":           "epdisc.ice.username",
		"password":           "epdisc.ice.password",
		"ice-candidate-type": "epdisc.ice.candidate_types",
		"ice-network-type":   "epdisc.ice.network_types",

		// Peer discovery
		"pdisc":     "pdisc.enabled",
		"community": "pdisc.community",
	}
)

type Config struct {
	Settings

	*Meta
	*koanf.Koanf

	Runtime *koanf.Koanf

	// Settings which are not configurable via configuration file
	Files                  []string
	Domains                []string
	DefaultInterfaceFilter string
	InterfaceOrder         []string

	mu     sync.Mutex
	flags  *pflag.FlagSet
	logger *zap.Logger
}

// ParseArgs creates a new configuration instance and loads all configuration
//
// Only used for testing.
func ParseArgs(args ...string) (*Config, error) {
	c := New(nil)

	if err := c.flags.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse command line flags: %w", err)
	}

	return c, c.Load()
}

// New creates a new configuration instance.
func New(flags *pflag.FlagSet) *Config {
	if flags == nil {
		flags = pflag.NewFlagSet("", pflag.ContinueOnError)
	}

	c := &Config{
		Koanf: koanf.NewWithConf(koanf.Conf{
			Delim: ".",
		}),
		Runtime: koanf.NewWithConf(koanf.Conf{
			Delim: ".",
		}),
		Meta:  Metadata(),
		flags: flags,
	}

	// Feature flags
	flags.BoolP("host-sync", "H", true, "Enable synchronization of /etc/hosts file")
	flags.BoolP("config-sync", "S", true, "Enable synchronization of WireGuard configuration files")
	flags.BoolP("endpoint-disc", "I", true, "Enable ICE endpoint discovery")
	flags.BoolP("route-sync", "R", true, "Enable synchronization of AllowedIPs and Kernel routing table")
	flags.BoolP("auto-config", "A", true, "Enable setup of link-local addresses and missing interface options")

	// Config flags
	flags.StringSliceVarP(&c.Domains, "domain", "D", []string{}, "A DNS `domain` name used for DNS auto-configuration")
	flags.StringSliceVarP(&c.Files, "config", "c", []string{}, "One or more `filename`s of configuration files")

	// Daemon flags
	flags.StringSliceP("backend", "b", []string{}, "One or more `URL`s to signaling backends")
	flags.DurationP("watch-interval", "i", 0, "An interval at which we are periodically polling the kernel for updates on WireGuard interfaces")

	// RPC socket flags
	flags.StringP("rpc-socket", "s", "", "The `path` of the unix socket used by other cunicu commands")
	flags.Bool("rpc-wait", false, "Wait until first client connected to control socket before continuing start")

	// WireGuard
	flags.StringVarP(&c.DefaultInterfaceFilter, "interface-filter", "f", "*", "A glob(7) `pattern` for filtering WireGuard interfaces which this daemon will manage (e.g. \"wg*\")")
	flags.BoolP("wg-userspace", "u", false, "Use user-space WireGuard implementation for newly created interfaces")

	// Config sync
	flags.StringP("config-path", "w", "", "The `directory` of WireGuard wg/wg-quick configuration files")
	flags.BoolP("config-watch", "W", false, "Watch and synchronize changes to the WireGuard configuration files")

	// Route sync
	flags.IntP("route-table", "T", DefaultRouteTable, "Kernel routing table to use")

	// Endpoint discovery
	flags.StringSliceP("url", "a", []string{}, "One or more `URL`s of STUN and/or TURN servers")
	flags.StringP("username", "U", "", "The `username` for STUN/TURN credentials")
	flags.StringP("password", "P", "", "The `password` for STUN/TURN credentials")

	flags.StringSlice("ice-candidate-type", []string{}, "Usable `candidate-type`s (one of host, srflx, prflx, relay)")
	flags.StringSlice("ice-network-type", []string{}, "Usable `network-type`s (one of udp4, udp6, tcp4, tcp6)")

	// Peer discovery
	flags.StringP("community", "x", "", "A `passphrase` shared with other peers in the same community")

	return c
}

// Load loads configuration settings from various sources
//
// Settings are loaded in the following order where the later overwrite the previous settings:
// - defaults
// - dns lookups
// - configuration files
// - environment variables
// - command line flags
func (c *Config) Load() error {
	// We cant to this in NewConfig since its called by init()
	// at which time the logging system is not initialized yet.
	c.logger = zap.L().Named("config")

	// Load default settings
	if err := c.Koanf.Load(ConfMapProvider(&DefaultSettings), nil); err != nil {
		return fmt.Errorf("failed to load default settings: %w", err)
	}

	c.InterfaceOrder = []string{c.DefaultInterfaceFilter}

	// Load settings from DNS lookups
	for _, domain := range c.Domains {
		p := LookupProvider(domain)
		if err := c.Koanf.Load(p, nil); err != nil {
			return fmt.Errorf("DNS auto-configuration failed: %w", err)
		}

		c.Files = append(c.Files, p.Files...)
	}

	// Search for config files
	if len(c.Files) == 0 {
		searchPaths := []string{"/etc", "/etc/cunicu"}
		if homeDir := os.Getenv("HOME"); homeDir != "" {
			searchPaths = append(searchPaths,
				filepath.Join(homeDir, ".config"),
				filepath.Join(homeDir, ".config", "cunicu"),
			)
		}

		for _, path := range append(searchPaths, ".") {
			fn := filepath.Join(path, "cunicu.yaml")
			if fi, err := os.Stat(fn); err == nil && !fi.IsDir() {
				c.Files = append(c.Files, fn)
			}
		}
	}

	// Load config files
	for _, f := range c.Files {
		u, err := url.Parse(f)
		if err != nil {
			return fmt.Errorf("failed to load config file: invalid URL: %w", err)
		}

		p := YAMLFileProvider(u)
		if err := c.Koanf.Load(p, nil); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
		c.InterfaceOrder = append(c.InterfaceOrder, p.InterfaceOrder...)
	}

	// Load environment variables
	envKeyMap := map[string]string{}
	for _, k := range c.Meta.Keys() {
		m := strings.ToUpper(k)
		e := envPrefix + strings.ReplaceAll(m, ".", "_")
		envKeyMap[e] = k
	}

	c.Koanf.Load(env.ProviderWithValue(envPrefix, ".", func(e, v string) (string, any) {
		k := envKeyMap[e]

		if p := strings.Split(v, ","); len(p) > 1 {
			return k, p
		}

		return k, v

	}), nil)

	c.Koanf.Load(posflag.ProviderWithFlag(c.flags, ".", c.Koanf, func(f *pflag.Flag) (string, any) {
		setting, ok := flagMap[f.Name]
		if !ok {
			return "", nil
		}

		return setting, posflag.FlagVal(c.flags, f)
	}), nil)

	intfs := map[string]any{}

	// Default settings
	if c.DefaultInterfaceFilter != "" && (len(c.flags.Args()) == 0 || c.DefaultInterfaceFilter != "*") {
		k := fmt.Sprintf("interfaces.%s", c.DefaultInterfaceFilter)
		intfs[k] = Map(DefaultInterfaceSettings)
	}

	// Add interfaces from command line
	for _, i := range c.flags.Args() {
		k := fmt.Sprintf("interfaces.%s", i)
		intfs[k] = map[string]any{}
	}

	// Load interfaces
	if err := c.Koanf.Load(confmap.Provider(intfs, "."), nil); err != nil {
		return fmt.Errorf("failed to load: %w", err)
	}

	return c.Unmarshal()
}

// Check performs plausibility checks on the provided configuration.
func (c *Config) Check() error {
	if len(c.DefaultInterfaceSettings.EndpointDisc.ICE.URLs) > 0 && len(c.DefaultInterfaceSettings.EndpointDisc.ICE.CandidateTypes) > 0 {
		needsURL := false
		for _, ct := range c.DefaultInterfaceSettings.EndpointDisc.ICE.CandidateTypes {
			if ct.CandidateType == ice.CandidateTypeRelay || ct.CandidateType == ice.CandidateTypeServerReflexive {
				needsURL = true
			}
		}

		if !needsURL {
			c.logger.Warn("Ignoring supplied ICE URLs as there are no selected candidate types which would use them")
			c.DefaultInterfaceSettings.EndpointDisc.ICE.URLs = nil
		}
	}

	if c.DefaultInterfaceSettings.WireGuard.ListenPortRange.Min > c.DefaultInterfaceSettings.WireGuard.ListenPortRange.Max {
		return fmt.Errorf("invalid settings: WireGuard minimal listen port (%d) must be smaller or equal than maximal port (%d)",
			c.DefaultInterfaceSettings.WireGuard.ListenPortRange.Min,
			c.DefaultInterfaceSettings.WireGuard.ListenPortRange.Max,
		)
	}

	return nil
}

// Update sets multiple settings in the provided map.
// See also Set().
func (c *Config) Update(sets map[string]any) error {
	err := c.Runtime.Load(confmap.Provider(sets, "."), nil)
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	c.mu.Lock()
	err = c.Koanf.Merge(c.Runtime)
	c.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to merge config: %w", err)
	}

	if err := c.Unmarshal(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	return nil
}

// Set sets a single setting to the provided value
// The key should provided in its dot-delimited form.
func (c *Config) Set(key string, value any) error {
	return c.Update(map[string]any{key: value})
}

// MarshalRuntime writes the runtime configuration in YAML format to the provided writer.
func (c *Config) MarshalRuntime(wr io.Writer) error {
	out, err := c.Runtime.Marshal(yaml.Parser())
	if err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	if _, err := wr.Write(out); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	return nil
}

// Marshal writes the configuration in YAML format to the provided writer.
func (c *Config) Marshal(wr io.Writer) error {
	out, err := c.Koanf.Marshal(yaml.Parser())
	if err != nil {
		return err
	}

	_, err = wr.Write(out)

	return err
}

// Unmarshal populates the settings struct from the Koanf settings
func (c *Config) Unmarshal() error {
	if err := c.UnmarshalWithConf("", nil, koanf.UnmarshalConf{
		DecoderConfig: decoderConfig(&c.Settings),
	}); err != nil {
		return fmt.Errorf("failed unmarshal settings: %w", err)
	}

	isGlobPattern := func(str string) bool {
		return strings.ContainsAny(str, "*?[]^")
	}

	for k, v := range c.Interfaces {
		if isGlobPattern(k) {
			v.Pattern = k
		} else {
			v.Name = k
		}
		c.Interfaces[k] = v
	}

	return c.Check()
}

// InterfaceSettings returns interface specific settings
// These settings are constructed by merging the settings of
// each interface section which matches the name.
// This behavior is quite similar to the OpenSSH client configuration file.
func (c *Config) InterfaceSettings(name string) (cfg *InterfaceSettings) {
	for _, i := range c.InterfaceOrder {
		icfg := c.Interfaces[i]
		if !icfg.Matches(name) {
			continue
		}

		if cfg == nil {
			copy := icfg
			cfg = &copy
		} else {
			mergo.Merge(cfg, icfg,
				mergo.WithOverride,
				mergo.WithSliceDeepCopy)
		}
	}

	if cfg != nil {
		cfg.Name = name
		cfg.Pattern = ""
	}

	return cfg
}

// InterfaceFilter checks if the provided interface name is matched by any configuration.
func (c *Config) InterfaceFilter(name string) bool {
	for _, icfg := range c.Interfaces {
		if icfg.Matches(name) {
			return true
		}
	}

	return false
}

// decoderConfig returns the mapstructure DecoderConfig which is used by cunicu
func decoderConfig(result any) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToIPHookFunc(),
			mapstructure.StringToIPNetHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
			hookDecodeHook,
		),
		IgnoreUntaggedFields: true,
		WeaklyTypedInput:     true,
		ErrorUnused:          true,
		ZeroFields:           false,
		Result:               result,
		TagName:              "koanf",
	}
}
