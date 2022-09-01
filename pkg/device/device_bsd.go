//go:build darwin || freebsd

package device

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"

	"go.uber.org/zap"
)

type BSDKernelDevice struct {
	created bool
	index   int
	logger  *zap.Logger
}

func NewKernelDevice(name string) (*BSDKernelDevice, error) {
	if err := exec.Command("ifconfig", "wg", "create", "name", name).Run(); err != nil {
		return nil, err
	}

	return FindKernelDevice(name)
}

func FindKernelDevice(name string) (*BSDKernelDevice, error) {
	i, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}

	return &BSDKernelDevice{
		created: false,
		index:   i.Index,
		logger: zap.L().Named("device").With(
			zap.String("dev", name),
			zap.String("type", "kernel")),
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
	return exec.Command("ifconfig", d.Name(), "destroy").Run()
}

func (d *BSDKernelDevice) AddAddress(ip *net.IPNet) error {
	return exec.Command("ifconfig", d.Name(), ip.String(), "alias").Run()
}

func (d *BSDKernelDevice) DeleteAddress(ip *net.IPNet) error {
	return exec.Command("ifconfig", d.Name(), ip.String(), "-alias").Run()
}

func (d *BSDKernelDevice) Index() int {
	return d.index
}

var mtuRegex = regexp.MustCompile(`(?m)mtu (\d+)`)

func (d *BSDKernelDevice) MTU() int {
	out, err := exec.Command("ifconfig", d.Name()).Output()
	if err != nil {
		return -1
	}

	mtuStr := mtuRegex.FindString(string(out))
	if mtuStr == "" {
		return -1
	}

	mtu, err := strconv.Atoi(mtuStr)
	if err != nil {
		return -1
	}

	return mtu
}

func (d *BSDKernelDevice) SetMTU(mtu int) error {
	return exec.Command("ifconfig", d.Name(), "mtu", fmt.Sprint(mtu)).Run()
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
