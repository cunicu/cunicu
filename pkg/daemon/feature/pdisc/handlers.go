package pdisc

import (
	"fmt"
	"time"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon/feature/hsync"
	pdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/pdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
)

func (pd *Interface) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	// Ignore interface which do not have a private key yet
	if !i.PrivateKey().IsSet() {
		return
	}

	// Only send an update if the private key changed.
	// There are currently no other attributes which would need to be re-announced
	if m.Is(core.InterfaceModifiedPrivateKey) {
		var pkOld *crypto.Key
		if skOld := crypto.Key(old.PrivateKey); skOld.IsSet() {
			pk := skOld.PublicKey()
			pkOld = &pk
		}

		if err := pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_PEER_UPDATE, pkOld); err != nil {
			pd.logger.Error("Failed to send peer description", zap.Error(err))
		}
	}
}

func (pd *Interface) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if d := msg.Peer; pd != nil {
		if i := pd.Daemon.InterfaceByPublicKey(kp.Theirs); i != nil {
			// Received our own peer description. Ignoring...
			return
		}

		pk, err := crypto.ParseKeyBytes(d.PublicKey)
		if err != nil {
			pd.logger.Error("Failed to parse public key", zap.Error(err))
			return
		}

		if pk != kp.Theirs {
			pd.logger.Error("Received a peer description for from a wrong peer")
			return
		}

		if err := pd.OnPeerDescription(d); err != nil {
			pd.logger.Error("Failed to handle peer description", zap.Error(err), zap.Any("pd", msg.Peer))
		}
	}
}

func (pd *Interface) OnPeerDescription(d *pdiscproto.PeerDescription) error {
	pd.logger.Debug("Received peer description", zap.Any("description", d))

	pk, err := crypto.ParseKeyBytes(d.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	if !pd.isAccepted(pk) {
		pd.logger.Warn("Ignoring non-whitelisted peer", zap.Any("peer", pk))
		return nil
	}

	cp := pd.Peers[pk]

	switch d.Change {
	case pdiscproto.PeerDescriptionChange_PEER_ADD:
		if cp != nil {
			pd.logger.Warn("Peer already exists. Updating it instead")
			d.Change = pdiscproto.PeerDescriptionChange_PEER_UPDATE
		}

	case pdiscproto.PeerDescriptionChange_PEER_UPDATE:
		if cp == nil {
			pd.logger.Warn("Peer does not exist exists yet. Adding it instead")
			d.Change = pdiscproto.PeerDescriptionChange_PEER_ADD
		}

	default:
		if cp == nil {
			pd.logger.Warn("Ignoring non-existing peer")
			return nil
		}
	}

	cfg := d.Config()

	pd.descs[pk] = d

	switch d.Change {
	case pdiscproto.PeerDescriptionChange_PEER_ADD:
		if err := pd.AddPeer(&cfg); err != nil {
			return fmt.Errorf("failed to add peer: %w", err)
		}

	case pdiscproto.PeerDescriptionChange_PEER_UPDATE:
		if d.PublicKeyNew != nil {
			// Remove old peer
			if err := pd.RemovePeer(pk); err != nil {
				return fmt.Errorf("failed to remove peer: %w", err)
			}

			// Re-add peer with new public key
			if err := pd.AddPeer(&cfg); err != nil {
				return fmt.Errorf("failed to add peer: %w", err)
			}
		} else {
			if err := pd.UpdatePeer(&cfg); err != nil {
				return fmt.Errorf("failed to remove peer: %w", err)
			}

			pd.ApplyDescription(cp)

			// Update hostname if it has been changed
			if f, ok := pd.Features["hsync"]; ok {
				hs := f.(*hsync.Interface)
				if err := hs.Sync(); err != nil {
					return fmt.Errorf("failed to sync hosts: %w", err)
				}
			}
		}

	case pdiscproto.PeerDescriptionChange_PEER_REMOVE:
		if err := pd.RemovePeer(pk); err != nil {
			return fmt.Errorf("failed to remove peer: %w", err)
		}
	}

	// Re-announce ourself in case this is a new peer we did not knew already
	if cp == nil {
		// TODO: Fix the race which requires the delay
		time.AfterFunc(1*time.Second, func() {
			if err := pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_PEER_ADD, nil); err != nil {
				pd.logger.Error("Failed to send peer description", zap.Error(err))
			}
		})
	}

	return nil
}

func (pd *Interface) OnPeerAdded(p *core.Peer) {
	pd.ApplyDescription(p)
}

func (pd *Interface) OnPeerRemoved(p *core.Peer) {}
