package device

import (
	"net"
	"strconv"

	"github.com/stv0g/cunicu/pkg/errors"
)

type WindowsKernelDevice struct {
}

func (d *WindowsKernelDevice) AddAddress(ip net.IPNet) error {
	d.logger.Debug("Add address", zap.String("addr", ip.String()))

	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	i.logger.Debug("Add route",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()))

	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) DeleteAddress(ip net.IPNet) error {
	i.logger.Debug("Delete address", zap.String("addr", ip.String()))

	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	i.logger.Debug("Delete route",
		zap.String("dst", dst.String()))

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
	d.logger.Debug("Set link MTU", zap.Int("mtu", mtu))

	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) SetUp() error {
	d.logger.Debug("Set link up")

	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) SetDown() error {
	i.logger.Debug("Set link down")

	return errors.ErrNotSupported
}

func (d *WindowsKernelDevice) Close() error {
	i.logger.Debug("Deleting kernel device")

	return nil
}

func DetectMTU(ip net.IP, fwmark int) (int, error) {
	// TODO: Thats just a guess
	return 1500, nil
}

func DetectDefaultMTU(fwmark int) (int, error) {
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
