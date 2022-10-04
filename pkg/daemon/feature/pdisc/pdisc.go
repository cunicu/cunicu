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

	peerFilter map[crypto.Key]bool
	peerNames  map[crypto.Key]string

	logger *zap.Logger
}

func New(i *daemon.Interface) (daemon.Feature, error) {
	if !i.Settings.PeerDisc.Enabled || !crypto.Key(i.Settings.PeerDisc.Community).IsSet() {
		return nil, nil
	}

	pd := &Interface{
		Interface:  i,
		peerFilter: map[crypto.Key]bool{},
		peerNames:  map[crypto.Key]string{},
		logger:     zap.L().Named("pdisc").With(zap.String("intf", i.Name())),
	}

	for _, k := range pd.Settings.PeerDisc.Whitelist {
		pd.peerFilter[crypto.Key(k)] = true
	}

	for _, k := range pd.Settings.PeerDisc.Blacklist {
		pd.peerFilter[crypto.Key(k)] = false
	}

	// Avoid sending a peer description if the interface does not have a private key yet
	if i.PublicKey().IsSet() {
		if err := pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_PEER_ADD, nil); err != nil {
			pd.logger.Error("Failed to send peer description", zap.Error(err))
		}
	}

	i.OnModified(pd)
	i.OnPeer(pd)

	return pd, nil
}

func (pd *Interface) Start() error {
	pd.logger.Info("Started peer discovery")

	// Subscribe to peer updates
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
	pk := pd.PublicKey()

	// Gather all allowed IPs for this interface
	allowedIPs := []*net.IPNet{}

	// Static addresses
	for _, addr := range pd.Settings.AutoConfig.Addresses {
		_, bits := addr.Mask.Size()
		addr.Mask = net.CIDRMask(bits, bits)

		allowedIPs = append(allowedIPs, &addr)
	}

	// Auto-generated prefixes
	for _, pfx := range pd.Settings.AutoConfig.Prefixes {
		addr := pk.IPAddress(pfx)
		_, bits := addr.Mask.Size()
		addr.Mask = net.CIDRMask(bits, bits)

		allowedIPs = append(allowedIPs, &addr)
	}

	// Other networks
	for _, netw := range pd.Settings.PeerDisc.Networks {
		allowedIPs = append(allowedIPs, &netw)
	}

	d := &pdiscproto.PeerDescription{
		Change:     chg,
		Name:       pd.Settings.PeerDisc.Name,
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
	if verdict, ok := pd.peerFilter[pk]; ok {
		return verdict
	}

	return true
}
