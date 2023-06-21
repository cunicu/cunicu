// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package pdisc implements peer discovery based on a shared community passphrase.
package pdisc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/log"
	proto "github.com/stv0g/cunicu/pkg/proto/core"
	pdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/pdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/types/slices"
)

var errFailedUpdatePublicKey = errors.New("can not change public key in non-update message")

var Get = daemon.RegisterFeature(New, 60) //nolint:gochecknoglobals

type Interface struct {
	*daemon.Interface

	filter map[crypto.Key]bool
	descs  map[crypto.Key]*pdiscproto.PeerDescription

	logger *log.Logger
}

func New(i *daemon.Interface) (*Interface, error) {
	if !i.Settings.DiscoverPeers || !crypto.Key(i.Settings.Community).IsSet() {
		return nil, daemon.ErrFeatureDeactivated
	}

	pd := &Interface{
		Interface: i,
		filter:    map[crypto.Key]bool{},
		descs:     map[crypto.Key]*pdiscproto.PeerDescription{},
		logger:    log.Global.Named("pdisc").With(zap.String("intf", i.Name())),
	}

	for _, k := range pd.Settings.Whitelist {
		pd.filter[k] = true
	}

	for _, k := range pd.Settings.Blacklist {
		pd.filter[k] = false
	}

	// Avoid sending a peer description if the interface does not have a private key yet
	if i.PrivateKey().IsSet() {
		if err := pd.sendPeerDescription(pdiscproto.PeerDescriptionChange_ADD, nil); err != nil {
			pd.logger.Error("Failed to send peer description", zap.Error(err))
		}
	}

	i.AddModifiedHandler(pd)
	i.AddPeerHandler(pd)

	return pd, nil
}

func (i *Interface) Start() error {
	i.logger.Info("Started peer discovery")

	// Subscribe to peer updates
	kp := &crypto.KeyPair{
		Ours:   crypto.Key(i.Settings.Community),
		Theirs: crypto.Key{},
	}
	if _, err := i.Daemon.Backend.Subscribe(context.Background(), kp, i); err != nil {
		return fmt.Errorf("failed to subscribe on peer discovery channel: %w", err)
	}

	return nil
}

func (i *Interface) Close() error {
	if err := i.sendPeerDescription(pdiscproto.PeerDescriptionChange_REMOVE, nil); err != nil {
		i.logger.Error("Failed to send peer description", zap.Error(err))
	}

	return nil
}

func (i *Interface) Description(cp *daemon.Peer) *pdiscproto.PeerDescription {
	if d, ok := i.descs[cp.PublicKey()]; ok {
		return d
	}

	return nil
}

func (i *Interface) sendPeerDescription(chg pdiscproto.PeerDescriptionChange, pkOld *crypto.Key) error {
	pk := i.PublicKey()

	// Gather all allowed IPs for this interface
	allowedIPs := []*net.IPNet{}

	// Static addresses
	for _, addr := range i.Settings.Addresses {
		addr := addr

		_, bits := addr.Mask.Size()
		addr.Mask = net.CIDRMask(bits, bits)

		allowedIPs = append(allowedIPs, &addr)
	}

	// Auto-generated prefixes
	for _, pfx := range i.Settings.Prefixes {
		pfx := pfx

		addr := pk.IPAddress(pfx)

		_, bits := addr.Mask.Size()
		addr.Mask = net.CIDRMask(bits, bits)

		allowedIPs = append(allowedIPs, &addr)
	}

	// Other networks
	for _, netw := range i.Settings.Networks {
		netw := netw
		allowedIPs = append(allowedIPs, &netw)
	}

	d := &pdiscproto.PeerDescription{
		Change:     chg,
		Name:       i.Settings.HostName,
		AllowedIps: slices.String(allowedIPs),
		BuildInfo:  buildinfo.BuildInfo(),
		Hosts:      map[string]*pdiscproto.PeerAddresses{},
	}

	for name, addrs := range i.Settings.ExtraHosts {
		daddrs := []*proto.IPAddress{}

		for _, addr := range addrs {
			daddr := proto.Address(addr.IP)
			daddrs = append(daddrs, daddr)
		}

		d.Hosts[name] = &pdiscproto.PeerAddresses{
			Addresses: daddrs,
		}
	}

	if name := i.Settings.HostName; name != "" {
		daddrs := []*proto.IPAddress{}
		for _, pfx := range i.Settings.Prefixes {
			addr := pk.IPAddress(pfx)
			daddrs = append(daddrs, proto.Address(addr.IP))
		}

		d.Hosts[name] = &pdiscproto.PeerAddresses{
			Addresses: daddrs,
		}
	}

	if pkOld != nil {
		if d.Change != pdiscproto.PeerDescriptionChange_UPDATE {
			return errFailedUpdatePublicKey
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
		Theirs: crypto.Key(i.Settings.Community).PublicKey(),
	}

	if err := i.Daemon.Backend.Publish(context.Background(), kp, msg); err != nil {
		return err
	}

	i.logger.Debug("Send peer description", zap.Reflect("description", d))

	return nil
}

func (i *Interface) isAccepted(pk crypto.Key) bool {
	if verdict, ok := i.filter[pk]; ok {
		return verdict
	}

	return true
}

func (i *Interface) ApplyDescription(cp *daemon.Peer) {
	if d, ok := i.descs[cp.PublicKey()]; ok {
		cp.Name = d.Name

		if hosts := d.Hosts; len(hosts) > 0 {
			cp.Hosts = map[string][]net.IP{}

			for name, addrs := range hosts {
				hs := []net.IP{}
				for _, addr := range addrs.Addresses {
					hs = append(hs, addr.Address())
				}

				cp.Hosts[name] = hs
			}
		}
	}
}
