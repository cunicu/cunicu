package autocfg

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/pdisc"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func (ac *Interface) OnInterfaceModified(i *core.Interface, old *wg.Device, mod core.InterfaceModifier) {
	// Update addresses in case the interface key has changed
	if mod&core.InterfaceModifiedPrivateKey != 0 {
		if oldSk := crypto.Key(old.PrivateKey); oldSk.IsSet() {
			oldPk := oldSk.PublicKey()
			if err := ac.RemoveAddresses(oldPk); err != nil {
				ac.logger.Error("Failed to remove old addresses", zap.Error(err))
			}
		}

		if newSk := i.PrivateKey(); newSk.IsSet() {
			newPk := newSk.PublicKey()
			if err := ac.AddAddresses(newPk); err != nil {
				ac.logger.Error("Failed to add new addresses", zap.Error(err))
			}
		}
	}
}

func (ac *Interface) OnPeerAdded(p *core.Peer) {
	logger := ac.logger.With(zap.String("peer", p.String()))

	// Check if peer has been created by peer discovery
	var hasDesc bool
	if f, ok := ac.Interface.Features["pdisc"]; ok {
		hasDesc = f.(*pdisc.Interface).Description(p) != nil
	} else {
		hasDesc = false
	}

	// Add AllowedIPs for peer if they are not added by the peer-discovery
	if !hasDesc {
		for _, q := range ac.Settings.Prefixes {
			ip := p.PublicKey().IPAddress(q)

			_, bits := ip.Mask.Size()
			ip.Mask = net.CIDRMask(bits, bits)

			if err := p.AddAllowedIP(ip); err != nil {
				logger.Error("Failed to add auto-generated address to AllowedIPs", zap.Error(err))
			}
		}
	}
}

func (ac *Interface) OnPeerRemoved(p *core.Peer) {}
