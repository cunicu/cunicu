// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package config defines, loads and parses project wide configuration settings from various sources
package config

import (
	"fmt"
	"io"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"cunicu.li/cunicu/pkg/log"
	"cunicu.li/cunicu/pkg/types"
)

type Config struct {
	*Settings
	*Meta
	*koanf.Koanf
	Runtime *runtimeSource
	Sources []Source

	// Settings which are not configurable via configuration file
	Files   []string
	Domains []string
	Watch   bool

	Providers         []koanf.Provider
	InterfaceOrder    []string
	InterfaceOrderCLI []string

	onInterfaceChanged map[string]*Meta

	flags  *pflag.FlagSet
	logger *log.Logger
}

// New creates a new configuration instance.
func New(flags *pflag.FlagSet) *Config {
	cfg := &Config{
		Meta:               Metadata(),
		Runtime:            newRuntimeSource(),
		onInterfaceChanged: map[string]*Meta{},
		flags:              flags,
	}

	// Generic flags
	flags.StringArrayP("option", "o", nil, "Set arbitrary options (example: --option watch_interval=5s)")

	// Feature flags
	flags.BoolP("discover-endpoints", "E", true, "Enable ICE endpoint discovery")
	flags.BoolP("discover-peers", "P", true, "Enable peer discovery")
	flags.BoolP("sync-config", "C", true, "Enable synchronization of configuration files")
	flags.BoolP("sync-hosts", "H", true, "Enable synchronization of /etc/hosts file")
	flags.BoolP("sync-routes", "R", true, "Enable synchronization of AllowedIPs with Kernel routes")

	// Config flags
	flags.StringSliceVarP(&cfg.Domains, "domain", "D", []string{}, "A DNS `domain` name used for DNS auto-configuration")
	flags.StringSliceVarP(&cfg.Files, "config", "c", []string{}, "One or more `filename`s of configuration files")
	flags.BoolVarP(&cfg.Watch, "watch-config", "w", false, "Watch configuration for changes and apply changes at runtime.")

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
	flags.BoolP("port-forwarding", "F", true, "Enabled in-kernel port-forwarding")

	// Peer discovery
	flags.StringP("community", "x", "", "A `passphrase` shared with other peers in the same community")
	flags.StringP("hostname", "n", "", "A `name` which identifies this peer")

	return cfg
}

func (c *Config) Init(args []string) (err error) {
	// We recreate the logger here, as the logger created
	// in New() was created in init() before the logging system
	// was initialized.
	c.logger = log.Global.Named("config")

	// Initialize some defaults configuration settings at runtime
	if err := InitDefaults(); err != nil {
		return fmt.Errorf("failed to initialize defaults: %w", err)
	}

	c.InterfaceOrderCLI = args
	c.Sources = nil

	// Construct list of config sources
	providers, err := c.getProviders()
	if err != nil {
		return err
	}

	for _, provider := range providers {
		if err := c.AddProvider(provider); err != nil {
			return err
		}
	}

	if err := c.AddSource(c.Runtime); err != nil {
		return err
	}

	if _, err = c.ReloadAllSources(); err != nil {
		return err
	}

	return nil
}

func (c *Config) AddProvider(provider koanf.Provider) error {
	return c.AddSource(&source{
		Provider: provider,
	})
}

func (c *Config) AddSource(source Source) error {
	if w, ok := source.(Watchable); c.Watch && ok {
		if err := w.Watch(func(_ any, _ error) {
			if _, err := c.reload(func(s Source) bool { return s == source }); err != nil {
				c.logger.Error("Failed to reload config", zap.Error(err))
			}
		}); err != nil {
			return fmt.Errorf("failed to watch for changes: %w", err)
		}
	}

	c.Sources = append(c.Sources, source)

	return nil
}

// Update sets multiple settings in the provided map.
func (c *Config) Update(sets map[string]any) (map[string]types.Change, error) {
	if err := c.Runtime.Update(sets); err != nil {
		return nil, err
	}

	if err := c.Runtime.Save(); err != nil {
		return nil, err
	}

	changes, err := c.reload(func(_ Source) bool { return false })
	if err != nil {
		return nil, err
	}

	return changes, nil
}

// Marshal writes the configuration in YAML format to the provided writer.
func (c *Config) Marshal(wr io.Writer) error {
	return marshal(c.Koanf, wr)
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

// InterfaceOrderByName returns a list of interface config sections which are used by a given interface.
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

// ReloadAllSources reloads all configuration sources.
func (c *Config) ReloadAllSources() (map[string]types.Change, error) {
	return c.reload(func(_ Source) bool { return true })
}

// ReloadSource reloads a specific configuration source or all of nil is passed.
func (c *Config) reload(filter func(s Source) bool) (map[string]types.Change, error) {
	var err error

	newKoanf := koanf.New(".")
	newOrder := c.InterfaceOrderCLI

	for _, s := range c.Sources {
		if filter(s) {
			if err := s.Load(); err != nil {
				return nil, err
			}
		}

		if err := newKoanf.Merge(s.Config()); err != nil {
			return nil, fmt.Errorf("failed to merge: %w", err)
		}

		newOrder = append(newOrder, s.Order()...)
	}

	if len(newOrder) == 0 {
		newOrder = append(newOrder, "*")
	}

	if c.Settings, err = unmarshal(newKoanf); err != nil {
		return nil, err
	}

	// Detect changes
	changes := map[string]types.Change{}
	if c.Koanf != nil {
		changes = types.DiffMap(c.Koanf.Raw(), newKoanf.Raw())
	}

	c.Koanf = newKoanf
	c.InterfaceOrder = newOrder

	// Invoke onChanged handlers
	for key, change := range changes {
		c.logger.Info("Configuration setting changed",
			zap.String("key", key),
			zap.Any("old", change.Old),
			zap.Any("new", change.New))

		if err := c.InvokeChangedHandlers(key, change); err != nil {
			return nil, err
		}
	}

	return changes, nil
}

// DecoderConfig returns the mapstructure DecoderConfig which is used by cunicu.
func DecoderConfig(result any) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			mapstructure.StringToIPHookFunc(),
			mapstructure.TextUnmarshallerHookFunc(),
			stringsDecodeHook,
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

// unmarshal unmarshals the passed Koanf instance to a Settings struct.
func unmarshal(k *koanf.Koanf) (*Settings, error) {
	s := &Settings{}

	d, err := mapstructure.NewDecoder(DecoderConfig(s))
	if err != nil {
		return nil, err
	}

	if err := d.Decode(k.Raw()); err != nil {
		return nil, err
	}

	return s, s.Check()
}

func marshal(k *koanf.Koanf, wr io.Writer) error {
	out, err := k.Marshal(yaml.Parser())
	if err != nil {
		return err
	}

	_, err = wr.Write(out)

	return err
}
