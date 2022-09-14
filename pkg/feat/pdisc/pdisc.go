// Package pdisc implements peer discovery based on a shared community passphrase.
package pdisc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/pkg/wg"

	pdiscproto "github.com/stv0g/cunicu/pkg/proto/feat/pdisc"
)

type PeerDiscovery struct {
	backend signaling.Backend
	watcher *watcher.Watcher

	community crypto.Key

	whitelistMap map[crypto.Key]any

	logger *zap.Logger
}

func New(w *watcher.Watcher, c *wgctrl.Client, b signaling.Backend, community string, whitelist []crypto.Key) *PeerDiscovery {
	pd := &PeerDiscovery{
		backend:   b,
		watcher:   w,
		community: crypto.GenerateKeyFromPassword(community),
		logger:    zap.L().Named("pdisc"),
	}

	if whitelist != nil {
		pd.whitelistMap = map[crypto.Key]any{}
		for _, k := range whitelist {
			pd.whitelistMap[crypto.Key(k)] = nil
		}
	}

	w.OnInterface(pd)

	return pd
}

func (pd *PeerDiscovery) Start() error {
	pd.logger.Info("Started peer discovery")

	// Subscribe to peer updates
	// TODO: Support per-interface communities

	kp := &crypto.KeyPair{
		Ours:   pd.community,
		Theirs: signaling.AnyKey,
	}
	if _, err := pd.backend.Subscribe(context.Background(), kp, pd); err != nil {
		return fmt.Errorf("failed to subscribe on peer discovery channel: %w", err)
	}

	return nil
}

func (pd *PeerDiscovery) Close() error {
	return nil
}

func (pd *PeerDiscovery) OnInterfaceAdded(i *core.Interface) {
	i.OnModified(pd)

	// Ignore interface which do not have a private key yet
	if !i.PrivateKey().IsSet() {
		return
	}

	if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_ADD, nil); err != nil {
		pd.logger.Error("Failed to send peer description", zap.Error(err))
	}
}

func (pd *PeerDiscovery) OnInterfaceRemoved(i *core.Interface) {
	// Ignore interface which do not have a private key yet
	if !i.PrivateKey().IsSet() {
		return
	}

	if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_REMOVE, nil); err != nil {
		pd.logger.Error("Failed to send peer description", zap.Error(err))
	}
}

func (pd *PeerDiscovery) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
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

		if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_UPDATE, pkOld); err != nil {
			pd.logger.Error("Failed to send peer description", zap.Error(err))
		}
	}
}

func (pd *PeerDiscovery) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if msg.Peer != nil {
		if i := pd.watcher.InterfaceByPublicKey(kp.Theirs); i != nil {
			// Received our own peer description. Ignoring...
			return
		}

		if err := pd.onPeerDescription(msg.Peer); err != nil {
			pd.logger.Error("Failed to handle peer description", zap.Error(err), zap.Any("pd", msg.Peer))
		}
	}
}

func (pd *PeerDiscovery) onPeerDescription(pdisc *pdiscproto.PeerDescription) error {
	pk, err := crypto.ParseKeyBytes(pdisc.PublicKey)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	if !pd.isWhitelisted(pk) {
		pd.logger.Warn("Ignoring non-whitelisted peer", zap.Any("peer", pk))
		return nil
	}

	p := pd.watcher.PeerByPublicKey(&pk)

	switch pdisc.Change {
	case pdiscproto.PeerDescriptionChange_PEER_ADD:
		if p != nil {
			pd.logger.Warn("Peer already exists. Updating it instead", zap.String("intf", p.Interface.Name()))
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
		if err := pd.watcher.ForEachInterface(func(i *core.Interface) error {
			return i.AddPeer(&cfg)
		}); err != nil {
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
			if err := pd.watcher.ForEachInterface(func(i *core.Interface) error {
				return pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_ADD, nil)
			}); err != nil {
				pd.logger.Error("Failed to send peer description", zap.Error(err))
			}
		})
	}

	return nil
}

func (pd *PeerDiscovery) sendPeerDescription(i *core.Interface, chg pdiscproto.PeerDescriptionChange, pkOld *crypto.Key) error {
	d, err := i.MarshalDescription(chg, pkOld)
	if err != nil {
		return fmt.Errorf("failed to generate peer description: %w", err)
	}

	msg := &signaling.Message{
		Peer: d,
	}

	kp := &crypto.KeyPair{
		Ours:   i.PrivateKey(),
		Theirs: pd.community.PublicKey(),
	}

	if err := pd.backend.Publish(context.Background(), kp, msg); err != nil {
		return err
	}

	pd.logger.Debug("Send peer description", zap.Any("description", d))

	return nil
}

func (pd *PeerDiscovery) isWhitelisted(pk crypto.Key) bool {
	if pd.whitelistMap == nil {
		return true
	} else if _, ok := pd.whitelistMap[pk]; ok {
		return true
	}

	return false
}
