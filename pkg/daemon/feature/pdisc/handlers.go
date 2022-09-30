package pdisc

import (
	"fmt"
	"time"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
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
	if msg.Peer != nil {
		if i := pd.Daemon.Watcher.InterfaceByPublicKey(kp.Theirs); i != nil {
			// Received our own peer description. Ignoring...
			return
		}

		if err := pd.OnPeerDescription(msg.Peer); err != nil {
			pd.logger.Error("Failed to handle peer description", zap.Error(err), zap.Any("pd", msg.Peer))
		}
	}
}

func (pd *Interface) OnPeerDescription(pdisc *pdiscproto.PeerDescription) error {
	pk, err := crypto.ParseKeyBytes(pdisc.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	if !pd.isAccepted(pk) {
		pd.logger.Warn("Ignoring non-whitelisted peer", zap.Any("peer", pk))
		return nil
	}

	p := pd.Daemon.PeerByPublicKey(&pk)

	switch pdisc.Change {
	case pdiscproto.PeerDescriptionChange_PEER_ADD:
		if p != nil {
			pd.logger.Warn("Peer already exists. Updating it instead")
			pdisc.Change = pdiscproto.PeerDescriptionChange_PEER_UPDATE
		}

	case pdiscproto.PeerDescriptionChange_PEER_UPDATE:
		if p == nil {
			pd.logger.Warn("Peer does not exist exists yet. Adding it instead")
			pdisc.Change = pdiscproto.PeerDescriptionChange_PEER_ADD
		}

	default:
		if p == nil {
			return fmt.Errorf("cant remove non-existing peer")
		}
	}

	cfg := pdisc.Config()

	switch pdisc.Change {
	case pdiscproto.PeerDescriptionChange_PEER_ADD:
		if err := pd.AddPeer(&cfg); err != nil {
			return fmt.Errorf("failed to add peer: %w", err)
		}

	case pdiscproto.PeerDescriptionChange_PEER_UPDATE:
		if pdisc.PublicKeyNew != nil {
			// Remove old peer
			if err := p.Interface.RemovePeer(pk); err != nil {
				return fmt.Errorf("failed to remove peer: %w", err)
			}

			// Re-add peer with new public key
			if err := p.Interface.AddPeer(&cfg); err != nil {
				return fmt.Errorf("failed to add peer: %w", err)
			}
		} else {
			if err := p.Interface.UpdatePeer(&cfg); err != nil {
				return fmt.Errorf("failed to remove peer: %w", err)
			}
		}

	case pdiscproto.PeerDescriptionChange_PEER_REMOVE:
		if err := p.Interface.RemovePeer(pk); err != nil {
			return fmt.Errorf("failed to remove peer: %w", err)
		}
	}

	// Re-announce ourself in case this is a new peer we did not knew already
	if p == nil {
		// TODO: Check if delay is really necessary
		time.AfterFunc(1*time.Second, func() {
			if err := pd.Daemon.ForEachInterface(func(i *daemon.Interface) error {
				return pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_PEER_ADD, nil)
			}); err != nil {
				pd.logger.Error("Failed to send peer description", zap.Error(err))
			}
		})
	}

	return nil
}
