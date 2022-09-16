// Package pdisc implements peer discovery based on a shared community passphrase.
package pdisc

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/pkg/wg"

	pdiscproto "github.com/stv0g/cunicu/pkg/proto/feat/pdisc"
)

type PeerDiscovery struct {
	backend signaling.Backend
	watcher *watcher.Watcher

	config *config.Config

	whitelistMap map[*core.Interface]map[crypto.Key]any

	logger *zap.Logger
}

func New(w *watcher.Watcher, c *wgctrl.Client, b signaling.Backend, cfg *config.Config) *PeerDiscovery {
	pd := &PeerDiscovery{
		backend:      b,
		watcher:      w,
		whitelistMap: map[*core.Interface]map[crypto.Key]any{},
		config:       cfg,
		logger:       zap.L().Named("pdisc"),
	}

	w.OnInterface(pd)

	return pd
}

func (pd *PeerDiscovery) Start() error {
	pd.logger.Info("Started peer discovery")

	// Subscribe to peer updates
	// TODO: Support per-interface communities

	kp := &crypto.KeyPair{
		Ours:   crypto.Key(pd.config.DefaultInterfaceSettings.PeerDisc.Community),
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

	cfg := pd.config.InterfaceSettings(i.Name())

	if cfg.PeerDisc.Whitelist != nil {
		pd.whitelistMap[i] = map[crypto.Key]any{}
		for _, k := range cfg.PeerDisc.Whitelist {
			pd.whitelistMap[i][crypto.Key(k)] = nil
		}
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

	// TODO: support per interface communities
	whitelisted := false
	for i := range pd.whitelistMap {
		if pd.isWhitelisted(i, pk) {
			whitelisted = true
			break
		}
	}

	if !whitelisted {
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
	icfg := pd.config.InterfaceSettings(i.Name())

	// Gather all allowed IPs for this interface
	allowedIPs := []net.IPNet{}
	allowedIPs = append(allowedIPs, icfg.AutoConfig.Addresses...)

	if icfg.AutoConfig.LinkLocalAddresses {
		allowedIPs = append(allowedIPs,
			i.PublicKey().IPv6Address(),
			i.PublicKey().IPv4Address(),
		)
	}

	// Only the /32 or /128 for local addresses
	for _, allowedIP := range allowedIPs {
		for i := range allowedIP.Mask {
			allowedIP.Mask[i] = 0xff
		}
	}

	// But networks are taken in full
	allowedIPs = append(allowedIPs, icfg.PeerDisc.Networks...)

	d := &pdiscproto.PeerDescription{
		Change:     chg,
		Hostname:   icfg.PeerDisc.Hostname,
		AllowedIps: util.StringSlice(allowedIPs),
		BuildInfo:  buildinfo.BuildInfo(),
	}

	if pkOld != nil {
		if d.Change != pdiscproto.PeerDescriptionChange_PEER_UPDATE {
			return fmt.Errorf("can not change public key in non-update message")
		}

		d.PublicKeyNew = i.PublicKey().Bytes()
		d.PublicKey = pkOld.Bytes()
	} else {
		d.PublicKey = i.PublicKey().Bytes()
	}

	msg := &signaling.Message{
		Peer: d,
	}

	kp := &crypto.KeyPair{
		Ours:   i.PrivateKey(),
		Theirs: crypto.Key(icfg.PeerDisc.Community).PublicKey(),
	}

	if err := pd.backend.Publish(context.Background(), kp, msg); err != nil {
		return err
	}

	pd.logger.Debug("Send peer description", zap.Any("description", d))

	return nil
}

func (pd *PeerDiscovery) isWhitelisted(i *core.Interface, pk crypto.Key) bool {
	if pd.whitelistMap == nil {
		return true
	} else if wl, ok := pd.whitelistMap[i]; ok {
		if _, ok := wl[pk]; ok {
			return true
		}
	}

	return false
}
