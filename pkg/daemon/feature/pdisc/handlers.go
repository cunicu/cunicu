// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package pdisc

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/daemon/feature/hsync"
	pdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/pdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/wg"
)

func (i *Interface) OnInterfaceModified(ci *daemon.Interface, old *wg.Interface, m daemon.InterfaceModifier) {
	// Ignore interface which do not have a private key yet
	if !ci.PrivateKey().IsSet() {
		return
	}

	// Only send an update if the private key changed.
	// There are currently no other attributes which would need to be re-announced
	if m.Is(daemon.InterfaceModifiedPrivateKey) {
		var pkOld *crypto.Key
		if skOld := crypto.Key(old.PrivateKey); skOld.IsSet() {
			pk := skOld.PublicKey()
			pkOld = &pk
		}

		if err := i.sendPeerDescription(pdiscproto.PeerDescriptionChange_UPDATE, pkOld); err != nil {
			i.logger.Error("Failed to send peer description", zap.Error(err))
		}
	}
}

func (i *Interface) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if d := msg.Peer; i != nil {
		if i := i.Daemon.InterfaceByPublicKey(kp.Theirs); i != nil {
			// Received our own peer description. Ignoring...
			return
		}

		pk, err := crypto.ParseKeyBytes(d.PublicKey)
		if err != nil {
			i.logger.Error("Failed to parse public key", zap.Error(err))
			return
		}

		if pk != kp.Theirs {
			i.logger.Error("Received a peer description for from a wrong peer")
			return
		}

		if err := i.OnPeerDescription(d); err != nil {
			i.logger.Error("Failed to handle peer description", zap.Error(err), zap.Any("pd", msg.Peer))
		}
	}
}

func (i *Interface) OnPeerDescription(d *pdiscproto.PeerDescription) error { //nolint:gocognit
	i.logger.Debug("Received peer description", zap.Reflect("description", d))

	pk, err := crypto.ParseKeyBytes(d.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	if !i.isAccepted(pk) {
		i.logger.Warn("Ignoring non-whitelisted peer", zap.Any("peer", pk))
		return nil
	}

	cp := i.Peers[pk]

	switch d.Change {
	case pdiscproto.PeerDescriptionChange_ADD:
		if cp != nil {
			i.logger.Warn("Peer already exists. Updating it instead")
			d.Change = pdiscproto.PeerDescriptionChange_UPDATE
		}

	case pdiscproto.PeerDescriptionChange_UPDATE:
		if cp == nil {
			i.logger.Warn("Peer does not exist exists yet. Adding it instead")
			d.Change = pdiscproto.PeerDescriptionChange_ADD
		}

	default:
		if cp == nil {
			i.logger.Warn("Ignoring non-existing peer")
			return nil
		}
	}

	cfg := d.Config()

	i.descs[pk] = d

	switch d.Change {
	case pdiscproto.PeerDescriptionChange_ADD:
		if err := i.AddPeer(&cfg); err != nil {
			return fmt.Errorf("failed to add peer: %w", err)
		}

	case pdiscproto.PeerDescriptionChange_UPDATE:
		if d.PublicKeyNew != nil {
			// Remove old peer
			if err := i.RemovePeer(pk); err != nil {
				return fmt.Errorf("failed to remove peer: %w", err)
			}

			// Re-add peer with new public key
			if err := i.AddPeer(&cfg); err != nil {
				return fmt.Errorf("failed to add peer: %w", err)
			}
		} else {
			if err := i.UpdatePeer(&cfg); err != nil {
				return fmt.Errorf("failed to remove peer: %w", err)
			}

			i.ApplyDescription(cp)

			// Update hostname if it has been changed
			if hs := hsync.Get(i.Interface); hs != nil {
				if err := hs.Sync(); err != nil {
					return fmt.Errorf("failed to sync hosts: %w", err)
				}
			}
		}

	case pdiscproto.PeerDescriptionChange_REMOVE:
		if err := i.RemovePeer(pk); err != nil {
			return fmt.Errorf("failed to remove peer: %w", err)
		}
	}

	// Re-announce ourself in case this is a new peer we did not knew already
	if cp == nil {
		// TODO: Fix the race which requires the delay
		time.AfterFunc(1*time.Second, func() {
			if err := i.sendPeerDescription(pdiscproto.PeerDescriptionChange_ADD, nil); err != nil {
				i.logger.Error("Failed to send peer description", zap.Error(err))
			}
		})
	}

	return nil
}

func (i *Interface) OnPeerAdded(p *daemon.Peer) {
	i.ApplyDescription(p)
}

func (i *Interface) OnPeerRemoved(_ *daemon.Peer) {}
