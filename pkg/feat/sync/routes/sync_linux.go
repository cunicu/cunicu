package routes

import (
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"riasc.eu/wice/pkg/device"
)

func (s *RouteSynchronization) syncKernel() {
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

func (s *RouteSynchronization) watchKernel() {
	rus := make(chan netlink.RouteUpdate)
	errs := make(chan error)

	if err := netlink.RouteSubscribeWithOptions(rus, nil, netlink.RouteSubscribeOptions{
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
		}
	}
}

func (s *RouteSynchronization) handleRouteUpdate(ru *netlink.RouteUpdate) {
	s.logger.Debug("Received netlink route update", zap.Any("update", ru))

	if ru.Protocol == device.RouteProtocol {
		s.logger.Debug("Ignoring gateway which was installed by ourself", zap.Any("gw", ru.Gw))
		return
	}

	if ru.Gw == nil {
		s.logger.Debug("Ignoring route with missing gateway")
		return
	}

	if ru.Gw.To16() == nil {
		s.logger.Debug("Ignoring non-IPv6 gateway", zap.Any("gw", ru.Gw))
		return
	}

	if !ru.Gw.IsLinkLocalUnicast() {
		s.logger.Debug("Ignoring non-link-local gateway", zap.Any("gw", ru.Gw))
		return
	}

	hash := *(*gwHashV6)(ru.Gw[8:])

	p, ok := s.gwMapV6[hash]
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
