package device

import (
	"bufio"
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

const (
	ipr2TablesFile = "/etc/iproute2/rt_tables"
)

type LinuxKernelDevice struct {
	link netlink.Link

	logger *zap.Logger
}

func NewKernelDevice(name string) (*LinuxKernelDevice, error) {
	link := &netlink.Wireguard{
		LinkAttrs: netlink.NewLinkAttrs(),
	}
	link.LinkAttrs.Name = name

	if err := netlink.LinkAdd(link); err != nil {
		return nil, fmt.Errorf("failed to create WireGuard interface: %w", err)
	}

	return &LinuxKernelDevice{
		link: link,
		logger: zap.L().Named("dev").With(
			zap.String("dev", name),
			zap.String("type", "kernel")),
	}, nil
}

func FindKernelDevice(name string) (Device, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get link details: %w", err)
	}

	return &LinuxKernelDevice{
		link:   link,
		logger: zap.L().Named("dev").With(zap.String("dev", name)),
	}, nil
}

func (d *LinuxKernelDevice) Close() error {
	d.logger.Debug("Deleting kernel device")

	if err := netlink.LinkDel(d.link); err != nil {
		return fmt.Errorf("failed to delete WireGuard device: %w", err)
	}

	return nil
}

func (d *LinuxKernelDevice) Name() string {
	return d.link.Attrs().Name
}

func (d *LinuxKernelDevice) Index() int {
	return d.link.Attrs().Index
}

func (d *LinuxKernelDevice) Flags() net.Flags {
	i, err := net.InterfaceByIndex(d.Index())
	if err != nil {
		panic(err)
	}

	return i.Flags
}

func (d *LinuxKernelDevice) MTU() int {
	var err error

	d.link, err = netlink.LinkByIndex(d.Index())
	if err != nil {
		panic(err)
	}

	return d.link.Attrs().MTU
}

func (d *LinuxKernelDevice) SetMTU(mtu int) error {
	d.logger.Debug("Set link MTU", zap.Int("mtu", mtu))

	return netlink.LinkSetMTU(d.link, mtu)
}

func (d *LinuxKernelDevice) SetUp() error {
	d.logger.Debug("Set link up")

	return netlink.LinkSetUp(d.link)
}

func (d *LinuxKernelDevice) SetDown() error {
	d.logger.Debug("Set link down")

	return netlink.LinkSetDown(d.link)
}

func (d *LinuxKernelDevice) AddAddress(ip net.IPNet) error {
	d.logger.Debug("Add address", zap.String("addr", ip.String()))

	addr := &netlink.Addr{
		IPNet: &ip,
		Flags: unix.IFA_F_PERMANENT,
	}

	if ip.IP.IsLinkLocalUnicast() || ip.IP.IsLinkLocalMulticast() {
		addr.Scope = unix.RT_SCOPE_LINK
	}

	return netlink.AddrAdd(d.link, addr)
}

func (d *LinuxKernelDevice) DeleteAddress(ip net.IPNet) error {
	d.logger.Debug("Delete address", zap.String("addr", ip.String()))

	addr := &netlink.Addr{
		IPNet: &ip,
	}

	return netlink.AddrDel(d.link, addr)
}

func (d *LinuxKernelDevice) AddRoute(dst net.IPNet, gw net.IP, table int) error {
	d.logger.Debug("Add route",
		zap.String("dst", dst.String()),
		zap.String("gw", gw.String()))

	route := &netlink.Route{
		LinkIndex: d.link.Attrs().Index,
		Dst:       &dst,
		Protocol:  RouteProtocol,
		Table:     table,
		Gw:        gw,
	}

	if err := netlink.RouteAdd(route); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	return nil
}

func (d *LinuxKernelDevice) DeleteRoute(dst net.IPNet, table int) error {
	d.logger.Debug("Delete route",
		zap.String("dst", dst.String()))

	route := &netlink.Route{
		LinkIndex: d.link.Attrs().Index,
		Dst:       &dst,
		Table:     table,
	}

	return netlink.RouteDel(route)
}

func DetectMTU(ip net.IP, fwmark int) (int, error) {
	// TODO: How do we use the correct fwmark here?
	rts, err := netlink.RouteGet(ip)
	if err != nil {
		return -1, fmt.Errorf("failed to get route: %w", err)
	}

	return mtuFromRoutes(rts)
}

func DetectDefaultMTU(fwmark int) (int, error) {
	// TODO: How do we use the correct fwmark here?
	flt := &netlink.Route{
		Dst: nil,
	}

	rts, err := netlink.RouteListFiltered(unix.AF_INET, flt, netlink.RT_FILTER_DST)
	if err != nil {
		return -1, fmt.Errorf("failed to get route: %w", err)
	}

	return mtuFromRoutes(rts)
}

// mtuFromRoutes calculates the smallest MTU of from a set of routes
// by looking first at the per-route MTU attributes and secondly at
// the default MTU of the link which is used by the route as next hop.
func mtuFromRoutes(rts []netlink.Route) (int, error) {
	if len(rts) == 0 {
		return -1, errors.New("no route to destination")
	}

	var err error
	var mtu int
	var links = map[int]netlink.Link{}
	var linkMTU = math.MaxInt
	var routeMTU = math.MaxInt

	for _, rt := range rts {
		if rt.MTU != 0 && rt.MTU < routeMTU {
			routeMTU = rt.MTU
		}

		if rt.LinkIndex >= 0 {
			link, ok := links[rt.LinkIndex]
			if !ok {
				link, err = netlink.LinkByIndex(rt.LinkIndex)
				if err != nil {
					return -1, fmt.Errorf("failed to get interface with index %d: %w", rt.LinkIndex, err)
				}

				links[rt.LinkIndex] = link
			}

			if link.Attrs().MTU != 0 && link.Attrs().MTU < linkMTU {
				linkMTU = link.Attrs().MTU
			}
		}
	}

	if mtu = routeMTU; mtu == math.MaxInt {
		if mtu = linkMTU; mtu == math.MaxInt {
			return -1, fmt.Errorf("no routes or interfaces found")
		}
	}

	return mtu, nil
}

func Table(str string) (int, error) {
	if f, err := os.OpenFile(ipr2TablesFile, os.O_RDONLY, 0644); err == nil {
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			line = strings.Split(line, "#")[0]
			fields := strings.Fields(line)

			if len(fields) < 2 {
				continue
			}

			if fields[1] == str {
				str = fields[0]
				break
			}
		}
	}

	i, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}

	return i, nil
}
