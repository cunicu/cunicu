// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

func (c *Config) flagProvider() koanf.Provider {
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
