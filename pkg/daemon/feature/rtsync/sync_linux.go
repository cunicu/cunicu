package rtsync

import (
	"errors"
	"fmt"
	"net/netip"
	"syscall"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

// removeKernel removes all routes from the kernel which target
// the peers link local addresses as their destination
// or have the peers address configured as the gateway.
func (rs *Interface) removeKernel(p *core.Peer) error {
	pk := p.PublicKey()

	link, err := netlink.LinkByIndex(rs.KernelDevice.Index())
	if err != nil {
		return fmt.Errorf("failed to find link: %w", err)
	}

	// Get all IPv4 and IPv6 routes on link
	routes := []netlink.Route{}
	for _, af := range []int{unix.AF_INET, unix.AF_INET6} {
		routesAF, err := netlink.RouteList(link, af)
		if err != nil {
			rs.logger.Error("Failed to get routes from kernel", zap.Error(err))
		}

		routes = append(routes, routesAF...)
	}

	for _, route := range routes {
		// Skip routes not in the desired table
		if route.Table != rs.Settings.RoutingTable {
			continue
		}

		// Skip default routes
		if route.Dst == nil {
			continue
		}

		ours := false
		for _, q := range rs.Settings.Prefixes {
			gw := pk.IPAddress(q)

			if route.Gw == nil {
				if gw.IP.Equal(route.Dst.IP) {
					ours = true
				}
			} else {
				if !gw.IP.Equal(route.Gw) {
					ours = true
				}
			}
		}

		if !ours {
			continue
		}

		if err := p.Interface.KernelDevice.DeleteRoute(*route.Dst, rs.Settings.RoutingTable); err != nil && !errors.Is(err, syscall.ESRCH) {
			rs.logger.Error("Failed to delete route", zap.Error(err))
		}
	}

	return nil
}

// syncKernel adds routes from the kernel routing table as new AllowedIPs to the respective peer
// based on the destination address of the route.
func (rs *Interface) syncKernel() error {
	link, err := netlink.LinkByIndex(rs.KernelDevice.Index())
	if err != nil {
		return fmt.Errorf("failed to find link: %w", err)
	}

	for _, af := range []int{unix.AF_INET6, unix.AF_INET} {
		routes, err := netlink.RouteList(link, af)
		if err != nil {
			return fmt.Errorf("failed to list routes from kernel: %w", err)
		}

		for _, route := range routes {
			if err := rs.handleRouteUpdate(&netlink.RouteUpdate{
				Type:  unix.RTM_NEWROUTE,
				Route: route,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

// watchKernel watches for added/removed routes in the kernel routing table and adds/removes AllowedIPs
// to the respective peers based on the destination address of the routes.
func (s *Interface) watchKernel() {
	rus := make(chan netlink.RouteUpdate)
	errs := make(chan error)

	if err := netlink.RouteSubscribeWithOptions(rus, s.stop, netlink.RouteSubscribeOptions{
		ErrorCallback: func(err error) {
			errs <- err
		},
	}); err != nil {
		s.logger.Error("Failed to subscribe to netlink route updates", zap.Error(err))
		return
	}

	for {
		select {
		case ru := <-rus:
			if err := s.handleRouteUpdate(&ru); err != nil {
				s.logger.Error("Failed to handle route update", zap.Error(err))
			}

		case err := <-errs:
			s.logger.Error("Failed to monitor kernel route updates", zap.Error(err))

		case <-s.stop:
			return
		}
	}
}

func (s *Interface) handleRouteUpdate(ru *netlink.RouteUpdate) error {
	logger := s.logger.WithOptions(log.WithVerbose(10))

	logger.Debug("Received netlink route update", zap.Any("update", ru))

	if ru.Table != s.Settings.RoutingTable {
		logger.Debug("Ignore route from another table")
		return nil
	}

	if ru.Protocol == device.RouteProtocol {
		logger.Debug("Ignoring route which was installed by ourself")
		return nil
	}

	if ru.Gw == nil {
		logger.Debug("Ignoring route with missing gateway")
		return nil
	}

	// TODO
	if !ru.Gw.IsLinkLocalUnicast() {
		logger.Debug("Ignoring non-link-local gateway", zap.Any("gw", ru.Gw))
		return nil
	}

	gw, ok := netip.AddrFromSlice(ru.Gw)
	if !ok {
		return fmt.Errorf("failed to get address from slice")
	}

	p, ok := s.gwMap[gw]
	if !ok {
		logger.Debug("Ignoring unknown gateway", zap.Any("gw", ru.Gw))
		return nil
	}

	logger = logger.With(zap.String("peer", p.String()))

	if ru.LinkIndex != p.Interface.KernelDevice.Index() {
		logger.Debug("Ignoring gateway due to interface mismatch", zap.Any("gw", ru.Gw))
		return nil
	}

	for _, aip := range p.AllowedIPs {
		if util.ContainsNet(&aip, ru.Dst) {
			logger.Debug("Ignoring route as it is already covered by the current AllowedIPs",
				zap.Any("allowed_ip", aip),
				zap.Any("dst", ru.Dst))
			return nil
		}
	}

	switch ru.Type {
	case unix.RTM_NEWROUTE:
		if err := p.AddAllowedIP(*ru.Dst); err != nil {
			return fmt.Errorf("failed to add allowed IP: %w", err)
		}

	case unix.RTM_DELROUTE:
		if err := p.RemoveAllowedIP(*ru.Dst); err != nil {
			return fmt.Errorf("failed to remove allowed IP: %w", err)
		}
	}

	return nil
}

func (s *Interface) Sync() error {
	for _, af := range []int{unix.AF_INET, unix.AF_INET6} {
		rts, err := netlink.RouteListFiltered(af, &netlink.Route{
			Table: s.Settings.RoutingTable,
		}, netlink.RT_FILTER_TABLE)
		if err != nil {
			return fmt.Errorf("failed to list routes: %w", err)
		}

		for _, rte := range rts {
			if err := s.handleRouteUpdate(&netlink.RouteUpdate{
				Route: rte,
				Type:  unix.RTM_NEWROUTE,
			}); err != nil {
				s.logger.Error("Failed to handle route update", zap.Error(err))
			}
		}
	}

	return nil
}
