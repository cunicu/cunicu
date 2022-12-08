package device

import (
	"net"
	"strconv"

	"go.uber.org/zap"
)

type WindowsKernelDevice struct {
	index int

	logger *zap.Logger
}

func (d *WindowsKernelDevice) AddAddress(ip net.IPNet) error {
	d.logger.Debug("Add address", zap.String("addr", ip.String()))

	return errNotSupported
}

func (d *WindowsKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	d.logger.Debug("Add route",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()))

	return errNotSupported
}

func (d *WindowsKernelDevice) DeleteAddress(ip net.IPNet) error {
	d.logger.Debug("Delete address", zap.String("addr", ip.String()))

	return errNotSupported
}

func (d *WindowsKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	d.logger.Debug("Delete route",
		zap.String("dst", dst.String()))

	return errNotSupported
}

func (d *WindowsKernelDevice) Index() int {
	return -1
}

func (d *WindowsKernelDevice) Flags() net.Flags {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		panic(err)
	}

	return i.Flags
}

func (d *WindowsKernelDevice) MTU() int {
	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return -1
}

func (d *WindowsKernelDevice) SetMTU(mtu int) error {
	d.logger.Debug("Set link MTU", zap.Int("mtu", mtu))

	// MTU is a route attribute which we need to adjust for all routes added for the interface
	return errNotSupported
}

func (d *WindowsKernelDevice) SetUp() error {
	d.logger.Debug("Set link up")

	return errNotSupported
}

func (d *WindowsKernelDevice) SetDown() error {
	d.logger.Debug("Set link down")

	return errNotSupported
}

func (d *WindowsKernelDevice) Close() error {
	d.logger.Debug("Deleting kernel device")

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
