package intf

import (
	"errors"
	"net"
	"os"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func (i *BaseInterface) AddAddress(ip *net.IPNet) error {
	link := &netlink.Wireguard{
		LinkAttrs: netlink.LinkAttrs{
			Index: i.index,
		},
	}

	addr := &netlink.Addr{
		IPNet: ip,
		Flags: unix.IFA_F_PERMANENT,
		Scope: unix.RT_SCOPE_LINK,
	}

	if err := netlink.AddrAdd(link, addr); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	return nil
}

func (i *BaseInterface) AddRoute(dst *net.IPNet) error {
	route := &netlink.Route{
		LinkIndex: i.index,
		Dst:       dst,
	}

	return netlink.RouteAdd(route)
}
