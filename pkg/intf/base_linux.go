package intf

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

func (i *BaseInterface) addAddress(ip *net.IPNet) error {
	link := &netlink.Wireguard{
		LinkAttrs: netlink.LinkAttrs{
			Name: i.Device.Name,
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

func (i *BaseInterface) addRoute(dst *net.IPNet) error {
	link, err := netlink.LinkByName(i.Device.Name)
	if err != nil {
		return fmt.Errorf("failed to get interface index: %w", err)
	}

	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       dst,
	}

	return netlink.RouteAdd(route)
}
