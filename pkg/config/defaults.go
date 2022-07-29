package config

const (
	DefaultSocketPath    = "/var/run/wice.sock"
	DefaultSocketAddress = ":8810"
	DefaultURL           = "stun:l.google.com:19302"

	// Ephemeral Port Range (RFC6056 Sect. 2.1)
	EphemeralPortMin = (1 << 15) + (1 << 14)
	EphemeralPortMax = (1 << 16) - 1

	WireGuardDefaultPort = 51820
)

var (
	DefaultBackends = []string{"grpc://wice.0l.de:443"}
)

func (c *Config) SetDefaults() {
	c.SetDefault("backends", DefaultBackends)
	c.SetDefault("watch_interval", "1s")
	c.SetDefault("socket.path", DefaultSocketPath)
	c.SetDefault("socket.address", DefaultSocketAddress)
	c.SetDefault("socket.wait", false)
	c.SetDefault("auto_config.enabled", true)
	c.SetDefault("config_sync.enabled", false)
	c.SetDefault("config_sync.path", "/etc/wireguard")
	c.SetDefault("config_sync.watch", false)
	c.SetDefault("route_sync.enabled", false)
	c.SetDefault("route_sync.table", "main")
	c.SetDefault("host_sync.enabled", true)
	c.SetDefault("endpoint_disc.enabled", true)
	c.SetDefault("endpoint_disc.ice.check_interval", "200ms")
	c.SetDefault("endpoint_disc.ice.keepalive_interval", "2s")
	c.SetDefault("endpoint_disc.ice.disconnected_timeout", "5s")
	c.SetDefault("endpoint_disc.ice.restart_timeout", "5s")
	c.SetDefault("endpoint_disc.ice.failed_timeout", "5s")
	c.SetDefault("endpoint_disc.ice.max_binding_requests", 7)
	c.SetDefault("endpoint_disc.ice.urls", []string{DefaultURL})
	c.SetDefault("endpoint_disc.ice.port.min", EphemeralPortMin)
	c.SetDefault("endpoint_disc.ice.port.max", EphemeralPortMax)
}
