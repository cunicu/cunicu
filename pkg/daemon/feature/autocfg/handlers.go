// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package autocfg

import (
	"net"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/pdisc"
	"github.com/stv0g/cunicu/pkg/wg"
)

func (i *Interface) OnInterfaceModified(ci *daemon.Interface, old *wg.Interface, mod daemon.InterfaceModifier) {
	// Update addresses in case the interface key has changed
	if mod&daemon.InterfaceModifiedPrivateKey != 0 {
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

func (i *Interface) OnPeerAdded(p *daemon.Peer) {
	logger := i.logger.With(zap.String("peer", p.String()))

	// Check if peer has been created by peer discovery
	hasDesc := false
	if i := pdisc.Get(i.Interface); i != nil {
		hasDesc = i.Description(p) != nil
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

func (i *Interface) OnPeerRemoved(_ *daemon.Peer) {}
