// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"github.com/knadh/koanf/maps"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

type flagOptionProvider struct {
	flags *pflag.FlagSet
}

func (c *Config) flagOptionProvider() koanf.Provider {
	return &flagOptionProvider{c.flags}
}

func (p *flagOptionProvider) Read() (map[string]any, error) {
	options, err := p.flags.GetStringArray("option")
	if err != nil {
		return nil, err
	}

	settings := map[string]any{}

	for _, option := range options {
		p := strings.SplitN(option, "=", 2)
		if len(p) != 2 {
			continue
		}

		key, value := p[0], p[1]

		if oldValue, ok := settings[key]; ok {
			switch oldValue := oldValue.(type) {
			case []string:
				settings[key] = append(oldValue, value)
			case string:
				settings[key] = []string{oldValue, value}
			}
		} else {
			settings[key] = value
		}
	}

	return maps.Unflatten(settings, "."), nil
}

func (p *flagOptionProvider) ReadBytes() ([]byte, error) {
	return nil, errNotImplemented
}

func (c *Config) flagProvider() koanf.Provider {
	// Map flags from the flags to Koanf settings
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
		"port-forwarding": "port_forwarding",

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
