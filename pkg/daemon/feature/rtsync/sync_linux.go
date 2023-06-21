// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rtsync

import (
	"errors"
	"fmt"
	"net/netip"
	"syscall"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/link"
	netx "github.com/stv0g/cunicu/pkg/net"
)

// removeKernel removes all routes from the kernel which target
// the peers link local addresses as their destination
// or have the peers address configured as the gateway.
func (i *Interface) removeKernel(p *daemon.Peer) error {
	pk := p.PublicKey()

	link, err := netlink.LinkByIndex(i.Index())
	if err != nil {
		return fmt.Errorf("failed to find link: %w", err)
	}

	// Get all IPv4 and IPv6 rts on link
	rts := []netlink.Route{}
	for _, af := range []int{unix.AF_INET, unix.AF_INET6} {
		rtsAf, err := netlink.RouteList(link, af)
		if err != nil {
			i.logger.Error("Failed to get routes from kernel", zap.Error(err))
		}

		rts = append(rts, rtsAf...)
	}

	for _, rt := range rts {
		// Skip routes not in the desired table
		if rt.Table != i.Settings.RoutingTable {
			continue
		}

		// Skip default routes
		if rt.Dst == nil {
			continue
		}

		ours := false
		for _, q := range i.Settings.Prefixes {
			gw := pk.IPAddress(q)

			if rt.Gw == nil {
				if gw.IP.Equal(rt.Dst.IP) {
					ours = true
				}
			} else {
				if !gw.IP.Equal(rt.Gw) {
					ours = true
				}
			}
		}

		if !ours {
			continue
		}

		if err := p.Interface.DeleteRoute(*rt.Dst, i.Settings.RoutingTable); err != nil && !errors.Is(err, syscall.ESRCH) {
			i.logger.Error("Failed to delete route", zap.Error(err))
		}
	}

	return nil
}

// syncKernel adds routes from the kernel routing table as new AllowedIPs to the respective peer
// based on the destination address of the route.
func (i *Interface) syncKernel() error {
	for _, af := range []int{unix.AF_INET, unix.AF_INET6} {
		rts, err := netlink.RouteListFiltered(af, &netlink.Route{
			Table:     i.Settings.RoutingTable,
			LinkIndex: i.Device.Index(),
		}, netlink.RT_FILTER_TABLE|netlink.RT_FILTER_OIF)
		if err != nil {
			return fmt.Errorf("failed to list routes: %w", err)
		}

		for _, rte := range rts {
			if err := i.handleRouteUpdate(&netlink.RouteUpdate{
				Route: rte,
				Type:  unix.RTM_NEWROUTE,
			}); err != nil {
				i.logger.Error("Failed to handle route update", zap.Error(err))
			}
		}
	}

	return nil
}

// watchKernel watches for added/removed routes in the kernel routing table and adds/removes AllowedIPs
// to the respective peers based on the destination address of the routes.
func (i *Interface) watchKernel() error {
	rus := make(chan netlink.RouteUpdate)
	errs := make(chan error)

	if err := netlink.RouteSubscribeWithOptions(rus, i.stop, netlink.RouteSubscribeOptions{
		ListExisting: true,
		ErrorCallback: func(err error) {
			errs <- err
		},
	}); err != nil {
		return fmt.Errorf("failed to subscribe to netlink route updates: %w", err)
	}

	for {
		select {
		case ru := <-rus:
			if err := i.handleRouteUpdate(&ru); err != nil {
				i.logger.Error("Failed to handle route update", zap.Error(err))
			}

		case err := <-errs:
			i.logger.Error("Failed to monitor kernel route updates", zap.Error(err))

		case <-i.stop:
			return nil
		}
	}
}

func (i *Interface) handleRouteUpdate(ru *netlink.RouteUpdate) error {
	i.logger.DebugV(10, "Received netlink route update", zap.Reflect("update", ru))

	if ru.Table != i.Settings.RoutingTable {
		i.logger.DebugV(10, "Ignore route from another table")
		return nil
	}

	if ru.Protocol == link.RouteProtocol {
		i.logger.DebugV(10, "Ignoring route which was installed by ourself")
		return nil
	}

	if ru.Gw == nil {
		i.logger.DebugV(10, "Ignoring route with missing gateway")
		return nil
	}

	gw, ok := netip.AddrFromSlice(ru.Gw)
	if !ok {
		panic("failed to get address from slice")
	}

	p, ok := i.gwMap[gw]
	if !ok {
		i.logger.DebugV(10, "Ignoring unknown gateway", zap.Any("gw", ru.Gw))
		return nil
	}

	logger := i.logger.With(zap.String("peer", p.String()))

	if ru.LinkIndex != p.Interface.Device.Index() {
		logger.DebugV(10, "Ignoring gateway due to interface mismatch", zap.Any("gw", ru.Gw))
		return nil
	}

	for _, aip := range p.AllowedIPs {
		aip := aip

		if netx.ContainsNet(&aip, ru.Dst) {
			logger.DebugV(10, "Ignoring route as it is already covered by the current AllowedIPs",
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
