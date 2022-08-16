package device

import (
	"net"
)

const (
	RouteProtocol = 98
)

type KernelDevice interface {
	Close() error
	Delete() error

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

func NewDevice(name string, user bool) (kernelDev Device, err error) {
	if user {
		kernelDev, err = NewUserDevice(name)
	} else {
		kernelDev, err = NewKernelDevice(name)
	}
	if err != nil {
		return
	}

	return kernelDev, nil
}
