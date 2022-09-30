// Package pdisc implements peer discovery based on a shared community passphrase.
package pdisc

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"

	pdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/pdisc"
)

func init() {
	daemon.Features["pdisc"] = &daemon.FeaturePlugin{
		New:         New,
		Description: "Peer discovery",
		Order:       60,
	}
}

type Interface struct {
	*daemon.Interface

	whitelistMap map[crypto.Key]any
	blacklistMap map[crypto.Key]any

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.PeerDisc.Enabled || !crypto.Key(i.Settings.PeerDisc.Community).IsSet() {
		return nil, nil
	}

	pd := &Interface{
		Interface:    i,
		whitelistMap: map[crypto.Key]any{},
		blacklistMap: map[crypto.Key]any{},
		logger:       zap.L().Named("pdisc").With(zap.String("intf", i.Name())),
	}

	for _, k := range pd.Settings.PeerDisc.Whitelist {
		pd.whitelistMap[crypto.Key(k)] = nil
	}

	for _, k := range pd.Settings.PeerDisc.Blacklist {
		pd.blacklistMap[crypto.Key(k)] = nil
	}

	if err := pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_PEER_ADD, nil); err != nil {
		pd.logger.Error("Failed to send peer description", zap.Error(err))
	}

	i.OnModified(pd)

	return pd, nil
}

func (pd *Interface) Start() error {
	pd.logger.Info("Started peer discovery")

	// Subscribe to peer updates
	// TODO: Support per-interface communities

	kp := &crypto.KeyPair{
		Ours:   crypto.Key(pd.Settings.PeerDisc.Community),
		Theirs: signaling.AnyKey,
	}
	if _, err := pd.Daemon.Backend.Subscribe(context.Background(), kp, pd); err != nil {
		return fmt.Errorf("failed to subscribe on peer discovery channel: %w", err)
	}

	return nil
}

func (pd *Interface) Close() error {
	if err := pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_PEER_REMOVE, nil); err != nil {
		pd.logger.Error("Failed to send peer description", zap.Error(err))
	}

	return nil
}

func (pd *Interface) sendPeerDescription(chg pdiscproto.PeerDescriptionChange, pkOld *crypto.Key) error {
	// Gather all allowed IPs for this interface
	allowedIPs := []net.IPNet{}
	allowedIPs = append(allowedIPs, pd.Settings.AutoConfig.Addresses...)

	if pd.Settings.AutoConfig.LinkLocalAddresses {
		allowedIPs = append(allowedIPs,
			pd.PublicKey().IPv6Address(),
			pd.PublicKey().IPv4Address(),
		)
	}

	// Only the /32 or /128 for local addresses
	for _, allowedIP := range allowedIPs {
		for i := range allowedIP.Mask {
			allowedIP.Mask[i] = 0xff
		}
	}

	// But networks are taken in full
	allowedIPs = append(allowedIPs, pd.Settings.PeerDisc.Networks...)

	d := &pdiscproto.PeerDescription{
		Change:     chg,
		Hostname:   pd.Settings.PeerDisc.Hostname,
		AllowedIps: util.SliceString(allowedIPs),
		BuildInfo:  buildinfo.BuildInfo(),
	}

	if pkOld != nil {
		if d.Change != pdiscproto.PeerDescriptionChange_PEER_UPDATE {
			return fmt.Errorf("can not change public key in non-update message")
		}

		d.PublicKeyNew = pd.PublicKey().Bytes()
		d.PublicKey = pkOld.Bytes()
	} else {
		d.PublicKey = pd.PublicKey().Bytes()
	}

	msg := &signaling.Message{
		Peer: d,
	}

	kp := &crypto.KeyPair{
		Ours:   pd.PrivateKey(),
		Theirs: crypto.Key(pd.Settings.PeerDisc.Community).PublicKey(),
	}

	if err := pd.Daemon.Backend.Publish(context.Background(), kp, msg); err != nil {
		return err
	}

	pd.logger.Debug("Send peer description", zap.Any("description", d))

	return nil
}

func (pd *Interface) isAccepted(pk crypto.Key) bool {
	if pd.whitelistMap != nil {
		if _, ok := pd.whitelistMap[pk]; !ok {
			return false
		}
	}

	if pd.blacklistMap != nil {
		if _, ok := pd.whitelistMap[pk]; ok {
			return false
		}
	}

	return true
}
