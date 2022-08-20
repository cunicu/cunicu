// Package routes synchronizes the kernel routing table with the AllowedIPs of each WireGuard peer
package routes

import (
	"errors"
	"net"
	"net/netip"
	"syscall"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/watcher"
)

type RouteSync struct {
	watcher *watcher.Watcher

	gwMap map[netip.Addr]*core.Peer

	stop chan struct{}

	logger *zap.Logger
}

func New(w *watcher.Watcher, table string) *RouteSync {
	s := &RouteSync{
		watcher: w,
		gwMap:   map[netip.Addr]*core.Peer{},
		stop:    make(chan struct{}),
		logger:  zap.L().Named("sync.routes"),
	}

	w.OnPeer(s)

	go s.watchKernel()

	return s
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
			zap.String("intf", p.Interface.Name()),
			zap.String("peer", p.String()),
		)
	}
}

func (rs *RouteSync) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	for _, dst := range ipsAdded {
		if err := p.Interface.KernelDevice.AddRoute(&dst); err != nil {
			rs.logger.Error("Failed to add route", zap.Error(err))
			continue
		}

		rs.logger.Info("Added new AllowedIP to kernel routing table",
			zap.String("dst", dst.String()),
			zap.String("intf", p.Interface.Name()),
			zap.Any("peer", p.PublicKey()))
	}

	for _, dst := range ipsRemoved {
		if err := p.Interface.KernelDevice.DeleteRoute(&dst); err != nil && !errors.Is(err, syscall.ESRCH) {
			rs.logger.Error("Failed to delete route", zap.Error(err))
			continue
		}

		rs.logger.Info("Remove vanished AllowedIP from kernel routing table",
			zap.String("dst", dst.String()),
			zap.String("intf", p.Interface.Name()),
			zap.Any("peer", p.PublicKey()))
	}
}
