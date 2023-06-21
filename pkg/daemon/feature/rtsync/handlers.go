// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package rtsync

import (
	"errors"
	"net"
	"net/netip"
	"syscall"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/daemon"
)

func (i *Interface) OnPeerAdded(p *daemon.Peer) {
	pk := p.PublicKey()

	for _, q := range i.Settings.Prefixes {
		gwn := pk.IPAddress(q)
		gw, ok := netip.AddrFromSlice(gwn.IP)
		if !ok {
			panic("failed to get address from slice")
		}

		i.gwMap[gw] = p
	}

	// Initial sync
	if err := i.syncKernel(); err != nil {
		i.logger.Error("Failed to synchronize kernel routing table", zap.Error(err))
	}

	p.AddModifiedHandler(i)
}

func (i *Interface) OnPeerRemoved(p *daemon.Peer) {
	pk := p.PublicKey()

	for _, q := range i.Settings.Prefixes {
		gwn := pk.IPAddress(q)
		gw, ok := netip.AddrFromSlice(gwn.IP)
		if !ok {
			panic("failed to get address from slice")
		}

		delete(i.gwMap, gw)
	}

	if err := i.removeKernel(p); err != nil {
		i.logger.Error("Failed to remove kernel routes for peer",
			zap.Error(err),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p),
		)
	}
}

func (i *Interface) OnPeerModified(p *daemon.Peer, _ *wgtypes.Peer, _ daemon.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	pk := p.PublicKey()

	// Determine peer gateway address by using the first IPv4 and IPv6 prefix
	var gwV4, gwV6 net.IP
	for _, q := range i.Settings.Prefixes {
		isV6 := q.IP.To4() == nil
		n := pk.IPAddress(q)
		if isV6 && gwV6 == nil {
			gwV6 = n.IP
		}

		if !isV6 && gwV4 == nil {
			gwV4 = n.IP
		}
	}

	for _, dst := range ipsAdded {
		var gw net.IP
		if isV6 := dst.IP.To4() == nil; isV6 {
			gw = gwV6
		} else {
			gw = gwV4
		}

		ones, bits := dst.Mask.Size()
		if gw != nil && ones == bits && dst.IP.Equal(gw) {
			gw = nil
		}

		if err := p.Interface.Device.AddRoute(dst, gw, i.Settings.RoutingTable); err != nil {
			i.logger.Error("Failed to add route", zap.Error(err))
			continue
		}

		i.logger.Info("Added new AllowedIP to kernel routing table",
			zap.String("dst", dst.String()),
			zap.Any("gw", gw.String()),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p))
	}

	for _, dst := range ipsRemoved {
		if err := p.Interface.Device.DeleteRoute(dst, i.Settings.RoutingTable); err != nil && !errors.Is(err, syscall.ESRCH) {
			i.logger.Error("Failed to delete route", zap.Error(err))
			continue
		}

		i.logger.Info("Remove vanished AllowedIP from kernel routing table",
			zap.String("dst", dst.String()),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p))
	}
}
