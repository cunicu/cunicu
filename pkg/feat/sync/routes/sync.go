package routes

import (
	"errors"
	"net"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/watcher"
)

type gwHashV4 [2]byte
type gwHashV6 [8]byte

type RouteSynchronization struct {
	watcher *watcher.Watcher

	gwMapV4 map[gwHashV4]*core.Peer
	gwMapV6 map[gwHashV6]*core.Peer

	logger *zap.Logger
}

func New(w *watcher.Watcher, table string) (*RouteSynchronization, error) {
	s := &RouteSynchronization{
		watcher: w,
		gwMapV4: map[gwHashV4]*core.Peer{},
		gwMapV6: map[gwHashV6]*core.Peer{},
		logger:  zap.L().Named("sync.routes"),
	}

	w.OnPeer(s)

	go s.watchKernel()

	return s, nil
}

func (s *RouteSynchronization) OnPeerAdded(p *core.Peer) {
	ipV4 := p.PublicKey().IPv4Address()
	hashV4 := *(*gwHashV4)(ipV4.IP[14:])

	ipV6 := p.PublicKey().IPv6Address()
	hashV6 := *(*gwHashV6)(ipV6.IP[8:])

	s.gwMapV4[hashV4] = p
	s.gwMapV6[hashV6] = p

	s.syncKernel() // Initial sync

	p.OnModified(s)
}

func (s *RouteSynchronization) OnPeerRemoved(p *core.Peer) {
	ipV4 := p.PublicKey().IPv4Address()
	hashV4 := *(*gwHashV4)(ipV4.IP[14:])

	ipV6 := p.PublicKey().IPv6Address()
	hashV6 := *(*gwHashV6)(ipV6.IP[8:])

	delete(s.gwMapV4, hashV4)
	delete(s.gwMapV6, hashV6)

	// TODO: do we also remove the routes from the kernel?
}

func (s *RouteSynchronization) OnPeerModified(p *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
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
		if err := p.Interface.KernelDevice.DeleteRoute(&dst); err != nil && !errors.Is(err, unix.ESRCH) {
			s.logger.Error("Failed to delete route", zap.Error(err))
			continue
		}

		s.logger.Info("Remove vanished AllowedIP from kernel routing table",
			zap.String("dst", dst.String()),
			zap.String("intf", p.Interface.Name()),
			zap.Any("peer", p.PublicKey()))
	}
}
