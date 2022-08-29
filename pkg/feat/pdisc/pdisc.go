// Package pdisc implements peer discovery based on a shared community passphrase.
package pdisc

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/watcher"
	"riasc.eu/wice/pkg/wg"

	pdiscproto "riasc.eu/wice/pkg/proto/feat/pdisc"
)

type PeerDiscovery struct {
	backend signaling.Backend
	watcher *watcher.Watcher

	community crypto.Key

	logger *zap.Logger
}

func New(w *watcher.Watcher, c *wgctrl.Client, b signaling.Backend, community string) *PeerDiscovery {
	pd := &PeerDiscovery{
		backend:   b,
		watcher:   w,
		community: crypto.GenerateKeyFromPassword(community),
		logger:    zap.L().Named("pdisc"),
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

	// Ignore interface which dont have a private key yet
	if !i.PrivateKey().IsSet() {
		return
	}

	if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_ADD, nil); err != nil {
		pd.logger.Error("Failed to send peer description", zap.Error(err))
	}
}

func (pd *PeerDiscovery) OnInterfaceRemoved(i *core.Interface) {
	// Ignore interface which dont have a private key yet
	if !i.PrivateKey().IsSet() {
		return
	}

	if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_REMOVE, nil); err != nil {
		pd.logger.Error("Failed to send peer description", zap.Error(err))
	}
}

func (pd *PeerDiscovery) OnInterfaceModified(i *core.Interface, old *wg.Device, m core.InterfaceModifier) {
	// Ignore interface which dont have a private key yet
	if !i.PrivateKey().IsSet() {
		return
	}

	// Only send an update if the private key changed.
	// There are currently no other attributes which would need to be reannounced
	if m.Is(core.InterfaceModifiedPrivateKey) {
		if skOld := crypto.Key(old.PrivateKey); skOld.IsSet() {
			pkOld := skOld.PublicKey()
			if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_UPDATE, &pkOld); err != nil {
				pd.logger.Error("Failed to send peer description", zap.Error(err))
			}
		}
	}
}

func (pd *PeerDiscovery) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	if msg.Peer != nil {
		if i := pd.watcher.Interfaces.ByPublicKey(kp.Theirs); i != nil {
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

	p := pd.watcher.PeerByKey(&pk)
	cfg := pdisc.Config()

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

	switch pdisc.Change {
	case pdiscproto.PeerDescriptionChange_PEER_ADD:
		for _, i := range pd.watcher.Interfaces {
			if err := i.AddPeer(&cfg); err != nil {
				return fmt.Errorf("failed to add peer: %w", err)
			}
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

	// Re-announce ourself in case this is a new peer we didnt knew already
	if p == nil {
		time.AfterFunc(1*time.Second, func() {
			for _, i := range pd.watcher.Interfaces {
				if err := pd.sendPeerDescription(i, pdiscproto.PeerDescriptionChange_PEER_ADD, nil); err != nil {
					pd.logger.Error("Failed to send peer description", zap.Error(err))
				}
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
