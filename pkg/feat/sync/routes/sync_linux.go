package routes

import (
	"errors"
	"net/netip"
	"syscall"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/device"
)

func (s *RouteSync) removeKernel(p *core.Peer) error {
	// TODO: Handle IPv4 routes

	pk := p.PublicKey()
	gwV4, _ := netip.AddrFromSlice(pk.IPv4Address().IP)
	gwV6, _ := netip.AddrFromSlice(pk.IPv6Address().IP)

	routes, err := netlink.RouteList(nil, unix.AF_INET6)
	if err != nil {
		s.logger.Error("Failed to get routes from kernel", zap.Error(err))
	}

	for _, route := range routes {
		gw, _ := netip.AddrFromSlice(route.Gw)
		if gwV4.Compare(gw) == 0 || gwV6.Compare(gw) == 0 {
			if err := p.Interface.KernelDevice.DeleteRoute(route.Dst); err != nil && !errors.Is(err, syscall.ESRCH) {
				s.logger.Error("Failed to delete route", zap.Error(err))
				continue
			}
		}
	}

	return nil
}

func (s *RouteSync) syncKernel() {
	routes, err := netlink.RouteList(nil, unix.AF_INET6)
	if err != nil {
		s.logger.Error("Failed to get routes from kernel", zap.Error(err))
	}

	for _, route := range routes {
		s.handleRouteUpdate(&netlink.RouteUpdate{
			Type:  unix.RTM_NEWROUTE,
			Route: route,
		})
	}
}

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
			s.handleRouteUpdate(&ru)

		case err := <-errs:
			s.logger.Error("Failed to monitor kernel route updates", zap.Error(err))

		case <-s.stop:
			return
		}
	}
}

func (s *RouteSync) handleRouteUpdate(ru *netlink.RouteUpdate) {
	s.logger.Debug("Received netlink route update", zap.Any("update", ru))

	if ru.Protocol == device.RouteProtocol {
		s.logger.Debug("Ignoring gateway which was installed by ourself", zap.Any("gw", ru.Gw))
		return
	}

	if ru.Gw == nil {
		s.logger.Debug("Ignoring route with missing gateway")
		return
	}

	if !ru.Gw.IsLinkLocalUnicast() {
		s.logger.Debug("Ignoring non-link-local gateway", zap.Any("gw", ru.Gw))
		return
	}

	gw, _ := netip.AddrFromSlice(ru.Gw)

	p, ok := s.gwMap[gw]
	if !ok {
		s.logger.Debug("Ignoring unknown gateway", zap.Any("gw", ru.Gw))
		return
	}

	if p.Interface.KernelDevice.Index() != ru.LinkIndex {
		s.logger.Debug("Ignoring gateway due to interface mismatch", zap.Any("gw", ru.Gw))
		return
	}

	switch ru.Type {
	case unix.RTM_NEWROUTE:
		if err := p.AddAllowedIP(ru.Dst); err != nil {
			s.logger.Error("Failed to add allowed IP", zap.Error(err))
		}

	case unix.RTM_DELROUTE:
		if err := p.RemoveAllowedIP(ru.Dst); err != nil {
			s.logger.Error("Failed to remove allowed IP", zap.Error(err))
		}
	}
}
