package routes

import (
	"errors"
	"fmt"
	"net/netip"
	"syscall"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/device"
)

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
		gw, ok := netip.AddrFromSlice(route.Gw)
		if !ok {
			return errors.New("failed to get address from slice")
		}

		if gwV4.Compare(gw) == 0 || gwV6.Compare(gw) == 0 {
			if err := p.Interface.KernelDevice.DeleteRoute(route.Dst); err != nil && !errors.Is(err, syscall.ESRCH) {
				s.logger.Error("Failed to delete route", zap.Error(err))
				continue
			}
		}
	}

	return nil
}

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
	// s.logger.Debug("Received netlink route update", zap.Any("update", ru))

	if ru.Protocol == device.RouteProtocol {
		// s.logger.Debug("Ignoring route which was installed by ourself")
		return nil
	}

	if ru.Gw == nil {
		// s.logger.Debug("Ignoring route with missing gateway")
		return nil
	}

	if !ru.Gw.IsLinkLocalUnicast() {
		// s.logger.Debug("Ignoring non-link-local gateway", zap.Any("gw", ru.Gw))
		return nil
	}

	gw, ok := netip.AddrFromSlice(ru.Gw)
	if !ok {
		return fmt.Errorf("failed to get address from slice")
	}

	p, ok := s.gwMap[gw]
	if !ok {
		// s.logger.Debug("Ignoring unknown gateway", zap.Any("gw", ru.Gw))
		return nil
	}

	if p.Interface.KernelDevice.Index() != ru.LinkIndex {
		// s.logger.Debug("Ignoring gateway due to interface mismatch", zap.Any("gw", ru.Gw))
		return nil
	}

	switch ru.Type {
	case unix.RTM_NEWROUTE:
		if err := p.AddAllowedIP(ru.Dst); err != nil {
			return fmt.Errorf("failed to add allowed IP: %w", err)
		}

	case unix.RTM_DELROUTE:
		if err := p.RemoveAllowedIP(ru.Dst); err != nil {
			return fmt.Errorf("failed to remove allowed IP: %w", err)
		}
	}

	return nil
}
