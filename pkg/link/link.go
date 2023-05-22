package link

import "net"

const (
	EthernetMTU = 1500
)

const (
	TypeWireGuard = "wireguard"
)

type Link interface {
	Close() error

	// Getter

	Name() string
	Index() int
	MTU() int
	Flags() net.Flags
	Type() string

	// Setter

	SetMTU(mtu int) error
	SetUp() error
	SetDown() error

	AddAddress(ip net.IPNet) error
	AddRoute(dst net.IPNet, gw net.IP, table int) error

	DeleteAddress(ip net.IPNet) error
	DeleteRoute(dst net.IPNet, table int) error
}
