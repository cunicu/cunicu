package autocfg

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func (ac *Interface) OnInterfaceModified(i *core.Interface, old *wg.Device, mod core.InterfaceModifier) {
	// Update addresses in case the interface key has changed
	if mod&core.InterfaceModifiedPrivateKey != 0 {
		oldPk := crypto.Key(old.PublicKey)
		newPk := i.PublicKey()

		if oldPk.IsSet() {
			for _, pfx := range ac.Settings.Prefixes {
				addr := oldPk.IPAddress(pfx)
				if err := ac.KernelDevice.DeleteAddress(addr); err != nil {
					ac.logger.Error("Failed to un-assign address",
						zap.String("address", addr.String()),
						zap.Error(err))
				}
			}
		}

		if newPk.IsSet() {
			for _, pfx := range ac.Settings.Prefixes {
				addr := newPk.IPAddress(pfx)
				if err := ac.KernelDevice.AddAddress(addr); err != nil {
					ac.logger.Error("Failed to assign address",
						zap.String("address", addr.String()),
						zap.Error(err))
				}
			}
		}
	}
}

func (ac *Interface) OnPeerAdded(p *core.Peer) {
	logger := ac.logger.With(zap.Any("peer", p.PublicKey()))

	// Add AllowedIPs for peer
	for _, q := range ac.Settings.Prefixes {
		ip := p.PublicKey().IPAddress(q)

		_, bits := ip.Mask.Size()
		ip.Mask = net.CIDRMask(bits, bits)

		if err := p.AddAllowedIP(ip); err != nil {
			logger.Error("Failed to add auto-generated address to AllowedIPs", zap.Error(err))
		}
	}
}

func (ac *Interface) OnPeerRemoved(p *core.Peer) {}
