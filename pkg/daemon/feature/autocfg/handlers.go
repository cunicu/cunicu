package autocfg

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func (a *Interface) OnInterfaceModified(i *core.Interface, old *wg.Device, mod core.InterfaceModifier) {
	// Update link-local addresses in case the interface key has changed
	if mod&core.InterfaceModifiedPrivateKey != 0 {
		oldPk := crypto.Key(old.PublicKey)
		newPk := i.PublicKey()

		if oldPk.IsSet() {
			if err := deleteAddresses(i.KernelDevice,
				oldPk.IPv4Address(),
				oldPk.IPv6Address(),
			); err != nil {
				a.logger.Error("Failed to delete link-local addresses", zap.Error(err))
			}
		}

		if newPk.IsSet() {
			if err := addAddresses(i.KernelDevice,
				newPk.IPv4Address(),
				newPk.IPv6Address(),
			); err != nil {
				a.logger.Error("Failed to assign link-local addresses", zap.Error(err))
			}
		}
	}
}

func (a *Interface) OnPeerAdded(p *core.Peer) {
	logger := a.logger.With(zap.Any("peer", p.PublicKey()))

	// Add default link-local address as allowed IP
	ipV4 := p.PublicKey().IPv4Address()
	ipV6 := p.PublicKey().IPv6Address()

	ipV4.Mask = net.CIDRMask(32, 32)
	ipV6.Mask = net.CIDRMask(128, 128)

	if err := p.AddAllowedIP(ipV4); err != nil {
		logger.Error("Failed to add link-local IPv4 address to AllowedIPs", zap.Error(err))
	}

	if err := p.AddAllowedIP(ipV6); err != nil {
		logger.Error("Failed to add link-local IPv6 address to AllowedIPs", zap.Error(err))
	}
}

func (a *Interface) OnPeerRemoved(p *core.Peer) {}
