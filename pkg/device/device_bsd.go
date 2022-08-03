//go:build darwin || freebsd

package device

import (
	"net"
	"os/exec"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/errors"
)

type BSDKernelDevice struct {
	created bool
	index   int
	logger  *zap.Logger
}

func NewKernelDevice(name string) (KernelDevice, error) {
	if err := exec.Command("ifconfig", "wg", "create", "name", name).Run(); err != nil {
		return nil, err
	}

	return FindDevice(name)
}

func FindDevice(name string) (KernelDevice, error) {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	return &BSDKernelDevice{
		created: false,
		index:   i.Index,
		logger:  zap.L().Named("device").With(zap.String("dev", name)),
	}, nil
}

func (d *BSDKernelDevice) Name() string {
	i, err := net.InterfaceByIndex(d.index)
	if err != nil {
		panic(err)
	}

	return i.Name
}

func (d *BSDKernelDevice) Close() error {
	return nil
}

func (d *BSDKernelDevice) AddAddress(ip *net.IPNet) error {

	return exec.Command("ifconfig", d.Name(), ip.IP.String(), "netmask", ip.Mask.String(), "alias").Run()
}

func (d *BSDKernelDevice) DeleteAddress(ip *net.IPNet) error {
	return exec.Command("ifconfig", d.Name(), ip.IP.String(), "netmask", ip.Mask.String(), "delete").Run()
}

func (d *BSDKernelDevice) AddRoute(dst *net.IPNet) error {
	return exec.Command("route", "add", "-net", dst.String(), "-interface", d.Name()).Run()
}

func (d *BSDKernelDevice) DeleteRoute(dst *net.IPNet) error {
	return exec.Command("route", "delete", "-net", dst.String(), "-interface", d.Name()).Run()

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
