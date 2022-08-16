package device

import (
	"net"

	"riasc.eu/wice/pkg/errors"
)

type WindowsKernelDevice struct {
}

func (d *WindowsKernelDevice) AddAddress(ip net.IPNet) error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) AddRoute(dst net.IPNet) error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) Index() int {
	return -1
}

func (d *WindowsKernelDevice) MTU() int {
	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return -1
}

func (d *WindowsKernelDevice) SetMTU(mtu int) error {
	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) SetUp() error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) SetDown() error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) Close() error {
	return nil
}

func (d *WindowsKernelDevice) Delete() error {
	return errors.ErrNotSupported
}
