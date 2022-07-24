package config

const (
	DefaultSocketPath    = "/var/run/wice.sock"
	DefaultSocketAddress = ":8810"
	DefaultURL           = "stun:l.google.com:19302"

	// Ephemeral Port Range (RFC6056 Sect. 2.1)
	EphemeralPortMin = (1 << 15) + (1 << 14)
	EphemeralPortMax = (1 << 16) - 1

	WireguardDefaultPort = 51820
)

var (
	DefaultBackends = []string{"grpc:?insecure=true"}
)
