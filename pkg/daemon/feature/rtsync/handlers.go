package rtsync

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/util"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func (rs *Interface) OnPeerAdded(p *core.Peer) {
	pk := p.PublicKey()

	gwV4, ok := netip.AddrFromSlice(pk.IPv4Address().IP)
	if !ok {
		panic(fmt.Errorf("failed to get address from slice: %s", pk.IPv4Address().IP))
	}

	gwV6, ok := netip.AddrFromSlice(pk.IPv6Address().IP)
	if !ok {
		panic(fmt.Errorf("failed to get address from slice: %s", pk.IPv6Address().IP))
	}

	rs.gwMap[gwV4] = p
	rs.gwMap[gwV6] = p

	rs.syncKernel() // Initial sync

	p.OnModified(rs)
}

func (rs *Interface) OnPeerRemoved(p *core.Peer) {
	pk := p.PublicKey()

	gwV4, ok := netip.AddrFromSlice(pk.IPv4Address().IP)
	if !ok {
		panic(fmt.Errorf("failed to get address from slice: %s", pk.IPv4Address().IP))
	}

	gwV6, ok := netip.AddrFromSlice(pk.IPv6Address().IP)
	if !ok {
		panic(fmt.Errorf("failed to get address from slice: %s", pk.IPv6Address().IP))
	}

	delete(rs.gwMap, gwV4)
	delete(rs.gwMap, gwV6)

	if err := rs.removeKernel(p); err != nil {
		rs.logger.Error("Failed to remove kernel routes for peer",
			zap.Error(err),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p),
		)
	}
}

func (rs *Interface) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	pk := p.PublicKey()

	for _, dst := range ipsAdded {
		var gwn net.IPNet
		if isV6 := dst.IP.To4() == nil; isV6 {
			gwn = pk.IPv6Address()
		} else {
			gwn = pk.IPv4Address()
		}

		var gw net.IP
		if !util.ContainsNet(&gwn, &dst) {
			gw = gwn.IP
		}

		if err := p.Interface.KernelDevice.AddRoute(dst, gw, rs.Settings.RouteSync.Table); err != nil {
			rs.logger.Error("Failed to add route", zap.Error(err))
			continue
		}

		rs.logger.Info("Added new AllowedIP to kernel routing table",
			zap.String("dst", dst.String()),
			zap.Any("gw", gw.String()),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p))
	}

	for _, dst := range ipsRemoved {
		if err := p.Interface.KernelDevice.DeleteRoute(dst, rs.Settings.RouteSync.Table); err != nil && !errors.Is(err, syscall.ESRCH) {
			rs.logger.Error("Failed to delete route", zap.Error(err))
			continue
		}

		rs.logger.Info("Remove vanished AllowedIP from kernel routing table",
			zap.String("dst", dst.String()),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p))
	}
}
