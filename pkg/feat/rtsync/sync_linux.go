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

// removeKernel removes all routes from the kernel which have the peers link-local address
// configured as their destination
func (s *RouteSync) removeKernel(p *core.Peer) error {
	pk := p.PublicKey()
	gwV4, ok1 := netip.AddrFromSlice(pk.IPv4Address().IP)
	gwV6, ok2 := netip.AddrFromSlice(pk.IPv6Address().IP)
	if !ok1 || !ok2 {
		return errors.New("failed to get address from slice")
	}

	routes, err := netlink.RouteList(nil, unix.AF_INET6)
	if err != nil {
		s.logger.Error("Failed to get routes from kernel", zap.Error(err))
	}

	for _, route := range routes {
		if route.Table != s.table {
			continue
		}

		gw, ok := netip.AddrFromSlice(route.Gw)
		if !ok {
			return errors.New("failed to get address from slice")
		}

		if gwV4.Compare(gw) == 0 || gwV6.Compare(gw) == 0 {
			if err := p.Interface.KernelDevice.DeleteRoute(*route.Dst, s.table); err != nil && !errors.Is(err, syscall.ESRCH) {
				s.logger.Error("Failed to delete route", zap.Error(err))
				continue
			}
		}
	}

	return nil
}

// syncKernel adds routes from the kernel routing table as new AllowedIPs to the respective peer
// based on the destination address of the route.
func (s *RouteSync) syncKernel() error {
	routes, err := netlink.RouteList(nil, unix.AF_INET6)
	if err != nil {
		return fmt.Errorf("failed to list routes from kernel: %w", err)
	}

	for _, route := range routes {
		if err := s.handleRouteUpdate(&netlink.RouteUpdate{
			Type:  unix.RTM_NEWROUTE,
			Route: route,
		}); err != nil {
			return err
		}
	}

	return nil
}

// watchKernel watches for added/removed routes in the kernel routing table and adds/removes AllowedIPs
// to the respective peers based on the destination address of the routes.
func (s *RouteSync) watchKernel() {
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

func (s *RouteSync) handleRouteUpdate(ru *netlink.RouteUpdate) error {
	logger := s.logger.WithOptions(log.WithVerbose(10))

	logger.Debug("Received netlink route update", zap.Any("update", ru))

	if ru.Table != s.table {
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

	logger = logger.With(zap.Any("peer", p))

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
		if err := p.RemoveAllowedIP(ru.Dst); err != nil {
			return fmt.Errorf("failed to remove allowed IP: %w", err)
		}
	}

	return nil
}

func (s *RouteSync) Sync() error {
	if err := s.syncFamily(unix.AF_INET); err != nil {
		return fmt.Errorf("failed to sync IPv4 routes: %w", err)
	}

	if err := s.syncFamily(unix.AF_INET6); err != nil {
		return fmt.Errorf("failed to sync IPv6 routes: %w", err)
	}

	return nil
}

func (s *RouteSync) syncFamily(family int) error {
	rts, err := netlink.RouteListFiltered(family, &netlink.Route{
		Table: s.table,
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

	return nil
}
