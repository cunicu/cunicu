package device

import (
	"net"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	RouteProtocol = 98
)

type Devices []*wgtypes.Device

func (devs *Devices) GetByName(name string) *wgtypes.Device {
	for _, dev := range *devs {
		if dev.Name == name {
			return dev
		}
	}

	return nil
}

type KernelDevice interface {
	Close() error

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
