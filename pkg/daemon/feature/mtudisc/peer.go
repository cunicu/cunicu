package mtudisc

import (
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/device"

	xerrors "github.com/stv0g/cunicu/pkg/errors"
)

type Peer struct {
	*core.Peer
	Interface *Interface

	MTU int

	// Connected UDP socket for Path MTU Discovery (PMTUD)
	// Only for Linux
	fd int

	logger *zap.Logger
}

func (i *Interface) NewPeer(cp *core.Peer) *Peer {
	p := &Peer{
		Peer:      cp,
		Interface: i,

		logger: zap.L().Named("mtudisc").With(
			zap.String("intf", cp.Interface.Name()),
			zap.String("peer", cp.String())),
	}

	cp.OnModified(p)

	if p.Endpoint != nil {
		var err error
		if p.MTU, err = p.DetectMTU(); err != nil {
			i.logger.Error("Failed to detect MTU for peer", zap.Error(err), zap.String("peer", cp.String()))
		} else if err := i.UpdateMTU(); err != nil {
			i.logger.Error("Failed to update MTU", zap.Error(err))
		}
	}

	return p
}

func (p *Peer) Close() error {
	return nil
}

func (p *Peer) OnPeerModified(cp *core.Peer, old *wgtypes.Peer, m core.PeerModifier, ipsAdded, ipsRemoved []net.IPNet) {
	var err error

	if m.Is(core.PeerModifiedEndpoint) && cp.Endpoint != nil {
		p.MTU, err = p.DetectMTU()
		if err != nil {
			p.logger.Error("Failed to detect MTU for peer", zap.Error(err), zap.String("peer", cp.String()))
		} else if err := p.Interface.UpdateMTU(); err != nil {
			p.logger.Error("Failed to update MTU", zap.Error(err))
		}
	}
}

func (p *Peer) DetectMTU() (int, error) {
	mtu, err := p.DiscoverPathMTU()
	if err != nil {
		if errors.Is(err, xerrors.ErrNotSupported) {
			p.logger.Warn("Platform does not support path MTU discovery. Falling back to local MTU detection")
			return device.DetectMTU(p.Endpoint.IP, p.Interface.FirewallMark)
		}

		return -1, fmt.Errorf("path MTU discovery failed: %w", err)
	}

	return mtu, nil
}
