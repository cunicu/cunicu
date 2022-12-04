package autocfg

import (
	"net"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/pdisc"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func (i *Interface) OnInterfaceModified(ci *core.Interface, old *wg.Device, mod core.InterfaceModifier) {
	// Update addresses in case the interface key has changed
	if mod&core.InterfaceModifiedPrivateKey != 0 {
		if oldSk := crypto.Key(old.PrivateKey); oldSk.IsSet() {
			oldPk := oldSk.PublicKey()
			if err := i.RemoveAddresses(oldPk); err != nil {
				i.logger.Error("Failed to remove old addresses", zap.Error(err))
			}
		}

		if newSk := ci.PrivateKey(); newSk.IsSet() {
			newPk := newSk.PublicKey()
			if err := i.AddAddresses(newPk); err != nil {
				i.logger.Error("Failed to add new addresses", zap.Error(err))
			}
		}
	}
}

func (i *Interface) OnPeerAdded(p *core.Peer) {
	logger := i.logger.With(zap.String("peer", p.String()))

	// Check if peer has been created by peer discovery
	hasDesc := false
	if f, ok := i.Interface.Features["pdisc"]; ok {
		if i, ok := f.(*pdisc.Interface); ok {
			hasDesc = i.Description(p) != nil
		}
	}

	// Add AllowedIPs for peer if they are not added by the peer-discovery
	if !hasDesc {
		for _, q := range i.Settings.Prefixes {
			ip := p.PublicKey().IPAddress(q)

			_, bits := ip.Mask.Size()
			ip.Mask = net.CIDRMask(bits, bits)

			if err := p.AddAllowedIP(ip); err != nil {
				logger.Error("Failed to add auto-generated address to AllowedIPs", zap.Error(err))
			}
		}
	}
}

func (i *Interface) OnPeerRemoved(p *core.Peer) {}
