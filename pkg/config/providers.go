package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/spf13/pflag"
)

var (
	envPrefix = "CUNICU_"

	// Map flags from the flags to to Koanf settings
	flagMap = map[string]string{
		// Config sync
		"cfgsync": "cfgsync.enabled",

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
		StructsProvider(&DefaultSettings, "koanf"),
		WireGuardProvider(),
	}

	// Load settings from DNS lookups
	for _, domain := range c.Domains {
		p := LookupProvider(domain)
		ps = append(ps, p)
	}

	// Search for config files
	if len(c.Files) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory")
		}

		for _, path := range []string{"/etc", "/etc/cunicu", cwd} {
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
			p = RemoteFileProvider(u)
		case "":
			p = LocalFileProvider(u)
		default:
			return nil, fmt.Errorf("unsupported scheme '%s' for config file", u.Scheme)
		}

		ps = append(ps, p)
	}

	// Add a runtime configuration file if it exists
	if fi, err := os.Stat(RuntimeConfigFile); err == nil && !fi.IsDir() {
		ps = append(ps,
			LocalFileProvider(&url.URL{
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
	return posflag.ProviderWithFlag(c.flags, ".", nil, func(f *pflag.Flag) (string, any) {
		setting, ok := flagMap[f.Name]
		if !ok {
			return "", nil
		}

		return setting, posflag.FlagVal(c.flags, f)
	})
}
