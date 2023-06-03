// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

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
		"ice-url":            "ice.urls",
		"ice-username":       "ice.username",
		"ice-password":       "ice.password",
		"ice-candidate-type": "ice.candidate_types",
		"ice-network-type":   "ice.network_types",
		"ice-relay-tcp":      "ice.relay_tcp",
		"ice-relay-tls":      "ice.relay_tls",

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
