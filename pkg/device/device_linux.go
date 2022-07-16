package device

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

type LinuxKernelDevice struct {
	// This device has been created by wice.
	// Indicates that the device should be deleted during shutdown.
	created bool

	link netlink.Link

	Logger *zap.Logger
}

func NewKernelDevice(name string) (KernelDevice, error) {
	link := &netlink.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	link.LinkAttrs.Name = name

	if err := netlink.LinkAdd(link); err != nil {
		return nil, fmt.Errorf("failed to create Wireguard interface: %w", err)
	}

	return &LinuxKernelDevice{
		created: true,
		link:    link,
		Logger:  zap.L().Named("device").With(zap.String("dev", name)),
	}, nil
}

func FindDevice(name string) (KernelDevice, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get link details: %w", err)
	}

	return &LinuxKernelDevice{
		created: false,
		link:    link,
		Logger:  zap.L().Named("device").With(zap.String("dev", name)),
	}, nil
}

func (i *LinuxKernelDevice) Close() error {
	if i.created {
		if err := i.Delete(); err != nil {
			return err
		}
	}

	return nil
}

func (i *LinuxKernelDevice) Name() string {
	return i.link.Attrs().Name
}

func (i *LinuxKernelDevice) Index() int {
	return i.link.Attrs().Index
}

func (i *LinuxKernelDevice) MTU() int {
	var err error

	i.link, err = netlink.LinkByIndex(i.Index())
	if err != nil {
		panic(err)
	}

	return i.link.Attrs().MTU
}

func (i *LinuxKernelDevice) Delete() error {
	i.Logger.Debug("Deleting kernel device")

	if err := netlink.LinkDel(i.link); err != nil {
		return fmt.Errorf("failed to delete Wireguard device: %w", err)
	}

	return nil
}

func (i *LinuxKernelDevice) SetMTU(mtu int) error {
	i.Logger.Debug("Set link MTU", zap.Int("mtu", mtu))
	return netlink.LinkSetMTU(i.link, mtu)
}

func (i *LinuxKernelDevice) SetUp() error {
	i.Logger.Debug("Set link up")
	return netlink.LinkSetUp(i.link)
}

func (i *LinuxKernelDevice) SetDown() error {
	i.Logger.Debug("Set link down")
	return netlink.LinkSetDown(i.link)
}

func (i *LinuxKernelDevice) AddAddress(ip *net.IPNet) error {
	i.Logger.Debug("Add address", zap.String("addr", ip.String()))

	addr := &netlink.Addr{
		IPNet: ip,
		Flags: unix.IFA_F_PERMANENT,
	}

	if ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() {
		addr.Scope = unix.RT_SCOPE_LINK
	}

	return netlink.AddrAdd(i.link, addr)
}

func (i *LinuxKernelDevice) DeleteAddress(ip *net.IPNet) error {
	i.Logger.Debug("Delete address", zap.String("addr", ip.String()))

	addr := &netlink.Addr{
		IPNet: ip,
	}

	return netlink.AddrDel(i.link, addr)
}

func (i *LinuxKernelDevice) AddRoute(dst *net.IPNet) error {
	i.Logger.Debug("Add route", zap.String("dst", dst.String()))

	route := &netlink.Route{
		LinkIndex: i.link.Attrs().Index,
		Dst:       dst,
		Protocol:  RouteProtocol,
	}

	if err := netlink.RouteAdd(route); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	return nil
}

func (i *LinuxKernelDevice) DeleteRoute(dst *net.IPNet) error {
	i.Logger.Debug("Delete route", zap.String("dst", dst.String()))

	route := &netlink.Route{
		LinkIndex: i.link.Attrs().Index,
		Dst:       dst,
	}

	return netlink.RouteDel(route)
}
