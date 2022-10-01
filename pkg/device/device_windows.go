package device

import (
	"net"
	"strconv"

	"github.com/stv0g/cunicu/pkg/errors"
)

type WindowsKernelDevice struct {
}

func (d *WindowsKernelDevice) AddAddress(ip net.IPNet) error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) DeleteAddress(ip net.IPNet) error {
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
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

func DetectMTU(ip net.IP) (int, error) {
	// TODO: Thats just a guess
	return 1500, nil
}

func DetectDefaultMTU() (int, error) {
	// TODO: Thats just a guess
	return 1500, nil
}

func Table(str string) (int, error) {
	i, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}

	return i, nil
}
