package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

const (
	envPrefix = "CUNICU_"
)

var errUnsupportedScheme = errors.New("unsupported scheme")

type Watchable interface {
	Watch(cb func(event interface{}, err error)) error
}

type Orderable interface {
	Order() []string
}

type SubProvidable interface {
	SubProviders() []koanf.Provider
}

type Versioned interface {
	Version() any
}

type Provider struct {
	koanf.Provider

	Config *koanf.Koanf
}

// Load loads configuration settings from various sources
//
// Settings are loaded in the following order where the later overwrite the previous settings:
// - defaults
// - dns lookups
// - configuration files
// - environment variables
// - command line flags
func (c *Config) GetProviders() ([]koanf.Provider, error) {
	ps := []koanf.Provider{
		NewStructsProvider(&DefaultSettings, "koanf"),
		NewWireGuardProvider(),
	}

	// Load settings from DNS lookups
	for _, domain := range c.Domains {
		p := NewLookupProvider(domain)
		ps = append(ps, p)
	}

	// Search for config files
	if len(c.Files) == 0 {
		searchPath := []string{"/etc", "/etc/cunicu"}

		if cwd, err := os.Getwd(); err != nil {
			c.logger.Warn("Failed to get working directory", zap.Error(err))
		} else {
			searchPath = append(searchPath, cwd)
		}

		if cfgDir := os.Getenv("CUNICU_CONFIG_DIR"); cfgDir != "" {
			searchPath = append(searchPath, cfgDir)
		}

		for _, path := range searchPath {
			fn := filepath.Join(path, "cunicu.yaml")
			if fi, err := os.Stat(fn); err == nil && !fi.IsDir() {
				c.Files = append(c.Files, fn)
			}
		}
	}

	// Add config files providers
	for _, f := range c.Files {
		u, err := url.Parse(f)
		if err != nil {
			return nil, fmt.Errorf("ignoring config file with invalid name: %w", err)
		}

		var p koanf.Provider
		switch u.Scheme {
		case "http", "https":
			p = NewRemoteFileProvider(u)
		case "":
			p = NewLocalFileProvider(u)
		default:
			return nil, fmt.Errorf("%w '%s' for config file", errUnsupportedScheme, u.Scheme)
		}

		ps = append(ps, p)
	}

	// Add a runtime configuration file if it exists
	if fi, err := os.Stat(RuntimeConfigFile); err == nil && !fi.IsDir() {
		ps = append(ps,
			NewLocalFileProvider(&url.URL{
				Path: RuntimeConfigFile,
			}),
		)
	}

	ps = append(ps,
		c.EnvironmentProvider(),
		c.FlagProvider(),
	)

	return ps, nil
}

func (c *Config) EnvironmentProvider() koanf.Provider {
	// Load environment variables
	envKeyMap := map[string]string{}
	for _, k := range c.Meta.Keys() {
		m := strings.ToUpper(k)
		e := envPrefix + strings.ReplaceAll(m, ".", "_")
		envKeyMap[e] = k
	}

	return env.ProviderWithValue(envPrefix, ".", func(e, v string) (string, any) {
		k := envKeyMap[e]

		if p := strings.Split(v, ","); len(p) > 1 {
			return k, p
		}

		return k, v
	})
}

func (c *Config) FlagProvider() koanf.Provider {
	// Map flags from the flags to to Koanf settings
	flagMap := map[string]string{
		// Feature flags
		"discover-peers":     "discover_peers",
		"discover-endpoints": "discover_endpoints",
		"sync-config":        "sync_config",
		"sync-hosts":         "sync_hosts",
		"sync-routes":        "sync_routes",

		"backend":        "backends",
		"watch-interval": "watch_interval",

		// Socket
		"rpc-socket": "rpc.socket",
		"rpc-wait":   "rpc.wait",

		// WireGuard
		"wg-userspace": "userspace",

		// Endpoint discovery
		"url":                "ice.urls",
		"username":           "ice.username",
		"password":           "ice.password",
		"ice-candidate-type": "ice.candidate_types",
		"ice-network-type":   "ice.network_types",
		"ice-relay-tcp":      "ice.relay_tcp",
		"ice-relay-tls":      "ice.relay_tls",

		// Peer discovery
		"community": "community",
		"hostname":  "hostname",

		// Route synchronization
		"routing-table": "routing_table",
	}

	return posflag.ProviderWithFlag(c.flags, ".", nil, func(f *pflag.Flag) (string, any) {
		setting, ok := flagMap[f.Name]
		if !ok {
			return "", nil
		}

		return setting, posflag.FlagVal(c.flags, f)
	})
}
