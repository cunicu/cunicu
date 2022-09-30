package wg

const (
	SocketPath = "/var/run/wireguard"
	ConfigPath = "/etc/wireguard"

	DefaultPort = 51820

	TunnelOverhead = 80 // Byte
	DefaultMTU     = 1500 - TunnelOverhead
	MinimalMTU     = 1280 // Byte for minimal IPv6 MTU
)
