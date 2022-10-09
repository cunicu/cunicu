// Package device implements OS abstractions for managing WireGuard links
package device

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

const (
	RouteProtocol = 98
)

type Device interface {
	Close() error

	// Getter

	Name() string
	Index() int
	MTU() int

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

func run(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	outStr := string(out)
	if err != nil {
		return "", fmt.Errorf("failed to run: %s\n%s", strings.Join(args, " "), outStr)
	}

	return outStr, nil
}
