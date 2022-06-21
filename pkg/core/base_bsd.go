//go:build darwin || dragonfly || freebsd || netbsd

package core

import (
	"net"
	"os/exec"
)

func (i *BaseInterface) AddAddress(ip *net.IPNet) error {
	return exec.Command("ifconfig", i.Device.Name, "alias", ip.String(), "up").Run()
}

func (i *BaseInterface) AddRoute(dst *net.IPNet) error {
	return exec.Command("route", "add", "-net", dst.String(), "-interface", i.Device.Name).Run()
}
