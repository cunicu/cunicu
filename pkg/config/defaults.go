package config

import "riasc.eu/wice/pkg/wg"

const (
	DefaultSocketPath = "/var/run/wice.sock"
	DefaultURL        = "stun:stun.l.google.com:19302"

	// Ephemeral Port Range (RFC6056 Sect. 2.1)
	EphemeralPortMin = (1 << 15) + (1 << 14)
	EphemeralPortMax = (1 << 16) - 1
)

var (
	DefaultBackends = []string{"grpc://wice.0l.de:443"}
)

func (c *Config) SetDefaults() {
	c.SetDefault("auto_config.enabled", true)
	c.SetDefault("backends", DefaultBackends)
	c.SetDefault("config_sync.enabled", true)
	c.SetDefault("config_sync.path", wg.ConfigPath)
	c.SetDefault("config_sync.watch", false)
	c.SetDefault("endpoint_disc.enabled", true)
	c.SetDefault("endpoint_disc.ice.check_interval", "200ms")
	c.SetDefault("endpoint_disc.ice.disconnected_timeout", "5s")
	c.SetDefault("endpoint_disc.ice.failed_timeout", "5s")
	c.SetDefault("endpoint_disc.ice.keepalive_interval", "2s")
	c.SetDefault("endpoint_disc.ice.max_binding_requests", 7)
	c.SetDefault("endpoint_disc.ice.port.max", EphemeralPortMax)
	c.SetDefault("endpoint_disc.ice.port.min", EphemeralPortMin)
	c.SetDefault("endpoint_disc.ice.restart_timeout", "5s")
	c.SetDefault("endpoint_disc.ice.urls", []string{DefaultURL})
	c.SetDefault("host_sync.enabled", true)
	c.SetDefault("route_sync.enabled", true)
	c.SetDefault("route_sync.table", "main")
	c.SetDefault("socket.path", DefaultSocketPath)
	c.SetDefault("socket.wait", false)
	c.SetDefault("watch_interval", "1s")
	c.SetDefault("wireguard.port.max", EphemeralPortMax)
	c.SetDefault("wireguard.port.min", wg.DefaultPort)
}
