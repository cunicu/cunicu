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

func New(w *watcher.Watcher, table string) (*RouteSync, error) {
	s := &RouteSync{
		watcher: w,
		gwMap:   map[netip.Addr]*core.Peer{},
		stop:    make(chan struct{}),
		logger:  zap.L().Named("sync.routes"),
	}

	w.OnPeer(s)

	go s.watchKernel()

	return s, nil
}

func (s *RouteSync) Close() error {
	// TODO: Remove Kernel routes added by us

	close(s.stop)

	return nil
}

func (s *RouteSync) OnPeerAdded(p *core.Peer) {
	pk := p.PublicKey()
	gwV4, _ := netip.AddrFromSlice(pk.IPv4Address().IP)
	gwV6, _ := netip.AddrFromSlice(pk.IPv6Address().IP)

	s.gwMap[gwV4] = p
	s.gwMap[gwV6] = p

	s.syncKernel() // Initial sync
}

func (s *RouteSync) OnPeerRemoved(p *core.Peer) {
	pk := p.PublicKey()
	gwV4, _ := netip.AddrFromSlice(pk.IPv4Address().IP)
	gwV6, _ := netip.AddrFromSlice(pk.IPv6Address().IP)

	delete(s.gwMap, gwV4)
	delete(s.gwMap, gwV6)

	if err := s.removeKernel(p); err != nil {
		s.logger.Error("Failed to remove kernel routes for peer",
			zap.Error(err),
			zap.String("intf", p.Interface.Name()),
			zap.String("peer", p.String()),
		)
	}
}

func (s *RouteSync) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	for _, dst := range ipsAdded {
		if err := p.Interface.KernelDevice.AddRoute(&dst); err != nil {
			s.logger.Error("Failed to add route", zap.Error(err))
			continue
		}

		s.logger.Info("Added new AllowedIP to kernel routing table",
			zap.String("dst", dst.String()),
			zap.String("intf", p.Interface.Name()),
			zap.Any("peer", p.PublicKey()))
	}

	for _, dst := range ipsRemoved {
		if err := p.Interface.KernelDevice.DeleteRoute(&dst); err != nil && !errors.Is(err, syscall.ESRCH) {
			s.logger.Error("Failed to delete route", zap.Error(err))
			continue
		}

		s.logger.Info("Remove vanished AllowedIP from kernel routing table",
			zap.String("dst", dst.String()),
			zap.String("intf", p.Interface.Name()),
			zap.Any("peer", p.PublicKey()))
	}
}
