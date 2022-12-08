// Package device implements OS abstractions for managing WireGuard links
package device

import (
	"errors"
	"net"
	"os"
)

const (
	RouteProtocol = 98
)

var errNotSupported = errors.New("not supported")

type Device interface {
	Close() error

	// Getter

	Name() string
	Index() int
	MTU() int
	Flags() net.Flags

	// Setter

	SetMTU(mtu int) error
	SetUp() error
	SetDown() error

	AddAddress(ip net.IPNet) error
	AddRoute(dst net.IPNet, gw net.IP, table int) error

	DeleteAddress(ip net.IPNet) error
	DeleteRoute(dst net.IPNet, table int) error
}

func NewDevice(name string, user bool) (kernelDev Device, err error) {
	if user {
		kernelDev, err = NewUserDevice(name)
	} else {
		kernelDev, err = NewKernelDevice(name)
	}
	if err != nil {
		return nil, err
	}

	return kernelDev, nil
}

func FindDevice(name string) (Device, error) {
	if dev, err := FindUserDevice(name); err == nil {
		return dev, nil
	} else if dev, err := FindKernelDevice(name); err == nil {
		return dev, nil
	}

	return nil, os.ErrNotExist
}
