// Package rtsync synchronizes the kernel routing table with the AllowedIPs of each WireGuard peer
package rtsync

import (
	"errors"
	"net"
	"net/netip"
	"syscall"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/watcher"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type RouteSync struct {
	gwMap map[netip.Addr]*core.Peer
	table int

	stop chan struct{}

	logger *zap.Logger
}

func New(w *watcher.Watcher, table int, watch bool) *RouteSync {
	rs := &RouteSync{
		gwMap:  map[netip.Addr]*core.Peer{},
		table:  table,
		stop:   make(chan struct{}),
		logger: zap.L().Named("rtsync"),
	}

	w.OnPeer(rs)

	if watch {
		go rs.watchKernel()
	}

	return rs
}

func (rs *RouteSync) Start() error {
	rs.logger.Info("Started route synchronization")

	return nil
}

func (rs *RouteSync) Close() error {
	// TODO: Remove Kernel routes added by us

	close(rs.stop)

	return nil
}

func (rs *RouteSync) OnPeerAdded(p *core.Peer) {
	pk := p.PublicKey()
	gwV4, ok1 := netip.AddrFromSlice(pk.IPv4Address().IP)
	gwV6, ok2 := netip.AddrFromSlice(pk.IPv6Address().IP)
	if !ok1 || !ok2 {
		panic("failed to get address from slice")
	}

	rs.gwMap[gwV4] = p
	rs.gwMap[gwV6] = p

	rs.syncKernel() // Initial sync
}

func (rs *RouteSync) OnPeerRemoved(p *core.Peer) {
	pk := p.PublicKey()
	gwV4, ok1 := netip.AddrFromSlice(pk.IPv4Address().IP)
	gwV6, ok2 := netip.AddrFromSlice(pk.IPv6Address().IP)
	if !ok1 || !ok2 {
		panic("failed to get address from slice")
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

func (rs *RouteSync) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	for _, dst := range ipsAdded {
		if err := p.Interface.KernelDevice.AddRoute(&dst, rs.table); err != nil {
			rs.logger.Error("Failed to add route", zap.Error(err))
			continue
		}

		rs.logger.Info("Added new AllowedIP to kernel routing table",
			zap.Any("dst", dst),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p))
	}

	for _, dst := range ipsRemoved {
		if err := p.Interface.KernelDevice.DeleteRoute(&dst, rs.table); err != nil && !errors.Is(err, syscall.ESRCH) {
			rs.logger.Error("Failed to delete route", zap.Error(err))
			continue
		}

		rs.logger.Info("Remove vanished AllowedIP from kernel routing table",
			zap.Any("dst", dst),
			zap.Any("intf", p.Interface),
			zap.Any("peer", p))
	}
}
