// Package mtudisc implements MTU discovery strategies
//
// We find a suitable MTU for the tunnel interface.
// The algorithm is the same as used by wg-quick:
//
//	The MTU is automatically determined from the endpoint addresses
//	or the system default route, which is usually a sane choice.

package mtudisc

import (
	"fmt"
	"math"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func init() {
	daemon.RegisterFeature("mtudisc", "MTU discovery", New, 11)
}

type Interface struct {
	*daemon.Interface

	defaultMTU int
	peers      map[*core.Peer]*Peer

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	m := &Interface{
		Interface: i,

		peers: map[*core.Peer]*Peer{},

		logger: zap.L().Named("mtudisc").With(zap.String("intf", i.Name())),
	}

	i.OnPeer(m)

	return m, nil
}

func (i *Interface) Start() error {
	var err error

	i.logger.Info("Started MTU discovery")

	// TODO: watch for changes of default route MTU
	if i.defaultMTU, err = device.DetectDefaultMTU(i.FirewallMark); err != nil {
		return err
	}

	if err := i.UpdateMTU(); err != nil {
		return fmt.Errorf("failed to update MTU: %w", err)
	}

	return nil
}

func (i *Interface) MTU() int {
	mtu := math.MaxInt
	for _, p := range i.peers {
		if p.MTU < mtu {
			mtu = p.MTU
		}
	}

	if mtu == math.MaxInt {
		mtu = i.defaultMTU
	}

	return mtu
}

func (i *Interface) TunnelMTU() int {
	return i.MTU() - wg.TunnelOverhead
}

func (i *Interface) UpdateMTU() error {
	newMTU := i.TunnelMTU()
	if newMTU < wg.MinimalMTU {
		return fmt.Errorf("MTU too small: %d", newMTU)
	}

	if oldMTU := i.KernelDevice.MTU(); newMTU != oldMTU { // Only update if changed
		if err := i.KernelDevice.SetMTU(newMTU); err != nil {
			return err
		}

		i.logger.Debug("Updated MTU of interface",
			zap.Int("new_mtu", newMTU),
			zap.Int("old_mtu", oldMTU))
	}

	return nil
}
