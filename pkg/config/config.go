// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package config defines, loads and parses project wide configuration settings from various sources
package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/imdario/mergo"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/stv0g/cunicu/pkg/log"
	"go.uber.org/zap"
)

type Config struct {
	*Settings
	*Meta
	*koanf.Koanf
	Runtime *koanf.Koanf
	Sources []*Source

	// Settings which are not configurable via configuration file
	Files   []string
	Domains []string
	Watch   bool

	Providers         []Provider
	InterfaceOrder    []string
	InterfaceOrderCLI []string

	onInterfaceChanged map[string]*Meta

	flags  *pflag.FlagSet
	logger *log.Logger
}

// ParseArgs creates a new configuration instance and loads all configuration
//
// Only used for testing.
func ParseArgs(args ...string) (*Config, error) {
	c := New(nil)

	if err := c.flags.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse command line flags: %w", err)
	}

	return c, c.Init(c.flags.Args())
}

// New creates a new configuration instance.
func New(flags *pflag.FlagSet) *Config {
	if flags == nil {
		flags = pflag.NewFlagSet("", pflag.ContinueOnError)
	}

	c := &Config{
		Meta:               Metadata(),
		Runtime:            koanf.New("."),
		onInterfaceChanged: map[string]*Meta{},
		flags:              flags,
	}

	// Feature flags
	flags.BoolP("discover-endpoints", "E", true, "Enable ICE endpoint discovery")
	flags.BoolP("discover-peers", "P", true, "Enable peer discovery")
	flags.BoolP("sync-config", "C", true, "Enable synchronization of configuration files")
	flags.BoolP("sync-hosts", "H", true, "Enable synchronization of /etc/hosts file")
	flags.BoolP("sync-routes", "R", true, "Enable synchronization of AllowedIPs with Kernel routes")

	// Config flags
	flags.StringSliceVarP(&c.Domains, "domain", "D", []string{}, "A DNS `domain` name used for DNS auto-configuration")
	flags.StringSliceVarP(&c.Files, "config", "c", []string{}, "One or more `filename`s of configuration files")
	flags.BoolVarP(&c.Watch, "watch-config", "w", false, "Watch configuration files for changes and apply changes at runtime.")

	// Daemon flags
	flags.StringSliceP("backend", "b", []string{}, "One or more `URL`s to signaling backends")
	flags.DurationP("watch-interval", "i", 0, "An interval at which we are periodically polling the kernel for updates on WireGuard interfaces")

	// RPC socket flags
	flags.StringP("rpc-socket", "s", "", "The `path` of the unix socket used by other cunicu commands")
	flags.Bool("rpc-wait", false, "Wait until first client connected to control socket before continuing start")

	// WireGuard
	flags.BoolP("wg-userspace", "U", false, "Use user-space WireGuard implementation for newly created interfaces")

	// Route sync
	flags.IntP("routing-table", "T", DefaultRouteTable, "Kernel routing table to use")

	// Endpoint discovery
	flags.StringSliceP("ice-url", "a", []string{}, "One or more `URL`s of STUN and/or TURN servers")
	flags.StringP("ice-username", "u", "", "The `username` for STUN/TURN credentials")
	flags.StringP("ice-password", "p", "", "The `password` for STUN/TURN credentials")

	flags.BoolP("port-forwarding", "F", true, "Enabled in-kernel port-forwarding")

	flags.StringSlice("ice-candidate-type", []string{}, "Usable `candidate-type`s (one of host, srflx, prflx, relay)")
	flags.StringSlice("ice-network-type", []string{}, "Usable `network-type`s (one of udp4, udp6, tcp4, tcp6)")

	flags.Bool("ice-relay-tcp", false, "Only use TCP relays")
	flags.Bool("ice-relay-tls", false, "Only use TLS secured relays")

	// Peer discovery
	flags.StringP("community", "x", "", "A `passphrase` shared with other peers in the same community")
	flags.StringP("hostname", "n", "", "A `name` which identifies this peer")

	return c
}

func (c *Config) Init(args []string) error {
	// We recreate the logger here, as the logger created
	// in New() was created in init() before the logging system
	// was initialized.
	c.logger = log.Global.Named("config")

	// Initialize some defaults configuration settings at runtime
	if err := InitDefaults(); err != nil {
		return fmt.Errorf("failed to initialize defaults: %w", err)
	}

	c.InterfaceOrderCLI = args

	ps, err := c.GetProviders()
	if err != nil {
		return err
	}

	for _, p := range ps {
		s := &Source{
			Provider: p,
		}

		c.Sources = append(c.Sources, s)

		if w, ok := p.(Watchable); c.Watch && ok {
			if err := w.Watch(func(event interface{}, err error) {
				if _, err := c.ReloadSource(s); err != nil {
					c.logger.Error("Failed to reload config", zap.Error(err))
				}
			}); err != nil {
				return fmt.Errorf("failed to watch for changes: %w", err)
			}
		}
	}

	_, err = c.Reload()

	return err
}

// Reload reloads all configuration sources
func (c *Config) Reload() (map[string]Change, error) {
	return c.ReloadSource(nil)
}

// ReloadSource reloads a specific configuration source or all of nil is passed
func (c *Config) ReloadSource(src *Source) (map[string]Change, error) {
	newKoanf := koanf.New(".")
	newOrder := c.InterfaceOrderCLI

	for _, s := range c.Sources {
		if src == nil || src == s {
			if err := s.Load(); err != nil {
				return nil, err
			}
		}

		if err := newKoanf.Merge(s.Config); err != nil {
			return nil, fmt.Errorf("failed to merge: %w", err)
		}

		newOrder = append(newOrder, s.Order...)
	}

	if err := newKoanf.Merge(c.Runtime); err != nil {
		return nil, err
	}

	if len(newOrder) == 0 {
		newOrder = append(newOrder, "*")
	}

	newSettings, err := Unmarshal(newKoanf)
	if err != nil {
		return nil, err
	}

	// Detect changes
	var changes map[string]Change
	if c.Koanf != nil {
		changes = DiffSettings(c.Settings, newSettings)
	}

	c.Settings = newSettings
	c.Koanf = newKoanf
	c.InterfaceOrder = newOrder

	// Invoke onChanged handlers
	for key, change := range changes {
		c.logger.Info("Configuration setting changed",
			zap.String("key", key),
			zap.Any("old", change.Old),
			zap.Any("new", change.New))

		c.InvokeHandlers(key, change)
	}

	return changes, nil
}

// Update sets multiple settings in the provided map.
func (c *Config) Update(sets map[string]any) (map[string]Change, error) {
	newRuntimeKoanf := c.Runtime.Copy()

	if err := newRuntimeKoanf.Load(confmap.Provider(sets, "."), nil); err != nil {
		return nil, err
	}

	newKoanf := c.Koanf.Copy()
	if err := newKoanf.Merge(newRuntimeKoanf); err != nil {
		return nil, err
	}

	newSettings, err := Unmarshal(newKoanf)
	if err != nil {
		return nil, err
	}

	// Detect changes
	var changes map[string]Change
	if c.Koanf != nil {
		changes = DiffSettings(c.Settings, newSettings)
	}

	c.Settings = newSettings
	c.Koanf = newKoanf
	c.Runtime = newRuntimeKoanf

	// Invoke onChanged handlers
	for key, change := range changes {
		c.logger.Info("Configuration setting changed",
			zap.String("key", key),
			zap.Any("old", change.Old),
			zap.Any("new", change.New))

		c.InvokeHandlers(key, change)
	}

	return changes, nil
}

// SaveRuntime saves the current runtime configuration to disk
func (c *Config) SaveRuntime() error {
	f, err := os.OpenFile(RuntimeConfigFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	defer f.Close()

	fmt.Fprintln(f, "# This is the cunīcu runtime configuration file.")
	fmt.Fprintln(f, "# It contains configuration adjustments made by")
	fmt.Fprintln(f, "# by the user with the cunicu-config-set(1) command.")
	fmt.Fprintln(f, "#")
	fmt.Fprintln(f, "# Please do not edit this file by hand as it will")
	fmt.Fprintln(f, "# be overwritten by cunīcu.")

	if len(c.Files) > 0 {
		fmt.Fprintln(f, "# Instead, please edit and more these settings")
		fmt.Fprintln(f, "# into the main configuration files:")
		for _, fn := range c.Files {
			fmt.Fprintf(f, "#  - %s\n", fn)
		}
	}

	fmt.Fprintln(f, "#")
	fmt.Fprintf(f, "# Last modification at %s\n", time.Now().Format(time.RFC1123Z))
	fmt.Fprintln(f, "---")

	return c.MarshalRuntime(f)
}

// MarshalRuntime writes the runtime configuration in YAML format to the provided writer.
func (c *Config) MarshalRuntime(wr io.Writer) error {
	if c.Runtime == nil {
		return nil
	}

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
	k := koanf.New(".")

	for _, src := range c.Sources {
		if err := k.Merge(src.Config); err != nil {
			return fmt.Errorf("failed to merge: %w", err)
		}
	}

	out, err := k.Marshal(yaml.Parser())
	if err != nil {
		return err
	}

	_, err = wr.Write(out)

	return err
}

// InterfaceSettings returns interface specific settings
// These settings are constructed by merging the settings of
// each interface section which matches the name.
// This behavior is quite similar to the OpenSSH client configuration file.
func (c *Config) InterfaceSettings(name string) (cfg *InterfaceSettings) {
	for _, set := range c.InterfaceOrderByName(name) {
		if cfg == nil {
			cfgCopy := c.DefaultInterfaceSettings
			cfg = &cfgCopy
		}

		if icfg, ok := c.Interfaces[set]; ok {
			if err := mergo.Merge(cfg, icfg, mergo.WithOverride); err != nil {
				panic(err)
			}
		}
	}

	return cfg
}

// InterfaceOrderByName returns a list of interface config sections which are used by a given interface
func (c *Config) InterfaceOrderByName(name string) []string {
	patterns := []string{}

	for _, pattern := range c.InterfaceOrder {
		if matched, err := filepath.Match(pattern, name); err == nil && matched {
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// InterfaceFilter checks if the provided interface name is matched by any configuration.
func (c *Config) InterfaceFilter(name string) bool {
	for _, pattern := range c.InterfaceOrder {
		if matched, err := filepath.Match(pattern, name); err == nil && matched {
			return true // Abort after first match
		}
	}

	return false
}

// DecoderConfig returns the mapstructure DecoderConfig which is used by cunicu
func DecoderConfig(result any) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToIPHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
			stringToIPAddrHook,
			stringToIPNetAddrHookFunc,
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

// Unmarshal unmarshals the passed Koanf instance to a Settings struct.
func Unmarshal(k *koanf.Koanf) (*Settings, error) {
	s := &Settings{}

	if err := k.UnmarshalWithConf("", nil, koanf.UnmarshalConf{
		DecoderConfig: DecoderConfig(s),
	}); err != nil {
		return nil, err
	}

	return s, s.Check()
}
