package device

import (
	"net"
)

const (
	RouteProtocol = 98
)

type KernelDevice interface {
	Close() error

	// Getter

	Name() string
	Index() int
	MTU() int

	// Setter

	SetMTU(mtu int) error
	SetUp() error
	SetDown() error

	AddAddress(ip *net.IPNet) error
	AddRoute(dst *net.IPNet) error

	DeleteAddress(ip *net.IPNet) error
	DeleteRoute(dst *net.IPNet) error
}
