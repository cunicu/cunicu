// go:build darwin || dragonfly || freebsd || netbsd

package device

import (
	"net"
	"os/exec"

	"riasc.eu/wice/pkg/errors"
)

type BSDKernelDevice struct {
	index int
}

func (d *BSDKernelDevice) AddAddress(ip net.IPNet) error {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		return err
	}

	return exec.Command("ifconfig", i.Name, "alias", ip.String(), "up").Run()
}

func (d *BSDKernelDevice) AddRoute(dst net.IPNet) error {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		return err
	}

	return exec.Command("route", "add", "-net", dst.String(), "-interface", i.Name).Run()
}

func (d *BSDKernelDevice) Index() int {
	return d.index
}

func (d *BSDKernelDevice) MTU() int {
	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return -1
}

func (d *BSDKernelDevice) SetMTU(mtu int) error {
	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return errors.ErrNotSupported
}

func (d *BSDKernelDevice) SetUp() error {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		return err
	}

	return exec.Command("ifconfig", "up", i.Name).Run()
}

func (d *BSDKernelDevice) SetDown() error {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		return err
	}

	return exec.Command("ifconfig", "down", i.Name).Run()
}
