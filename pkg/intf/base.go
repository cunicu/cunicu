package intf

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/internal/wg"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"
)

type BaseInterface struct {
	wgtypes.Device

	peers map[crypto.Key]Peer

	lastSync time.Time

	backend signaling.Backend
	args    *args.Args
	client  *wgctrl.Client
	server  *socket.Server

	logger *log.Entry
}

func (i *BaseInterface) Name() string {
	return i.Device.Name
}

func (i *BaseInterface) PublicKey() crypto.Key {
	return crypto.Key(i.Device.PublicKey)
}

func (i *BaseInterface) PrivateKey() crypto.Key {
	return crypto.Key(i.Device.PrivateKey)
}

func (i *BaseInterface) Close() error {
	i.logger.Info("Closing interface")

	for _, p := range i.peers {
		if err := p.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (i *BaseInterface) DumpConfig(wr io.Writer) {
	fmt.Fprintf(wr, "[Interface] # %s\n", i.Name())
	if i.PublicKey().IsSet() {
		fmt.Fprintf(wr, "PublicKey = %s\n", i.PublicKey())
	}
	if i.PrivateKey().IsSet() {
		fmt.Fprintf(wr, "PrivateKey = %s\n", i.PrivateKey())
	}
	if i.ListenPort != 0 {
		fmt.Fprintf(wr, "ListenPort = %d\n", i.ListenPort)
	}
	if i.FirewallMark != 0 {
		fmt.Fprintf(wr, "FwMark = %#x\n", i.FirewallMark)
	}

	for _, p := range i.Peers {
		fmt.Fprintf(wr, "[Peer]\n")
		if crypto.Key(p.PublicKey).IsSet() {
			fmt.Fprintf(wr, "PublicKey = %s\n", p.PublicKey)
		}
		if crypto.Key(p.PresharedKey).IsSet() {
			fmt.Fprintf(wr, "PresharedKey = %s\n", p.PresharedKey)
		}
		if !p.LastHandshakeTime.Equal(time.Time{}) {
			fmt.Fprintf(wr, "LastHandshakeTime = %v\n", p.LastHandshakeTime)
		}
		if p.PersistentKeepaliveInterval.Seconds() != 0 {
			fmt.Fprintf(wr, "PersistentKeepalive = %d # seconds\n", int(p.PersistentKeepaliveInterval.Seconds()))
		}
		if len(p.AllowedIPs) > 0 {
			aIPs := []string{}
			for _, aIP := range p.AllowedIPs {
				aIPs = append(aIPs, aIP.String())
			}
			fmt.Fprintf(wr, "AllowedIPs = %s\n", strings.Join(aIPs, ", "))
		}
		if p.Endpoint != nil {
			fmt.Fprintf(wr, "Endpoint = %s\n", p.Endpoint.String())
		}
	}
}

func (i *BaseInterface) syncPeer(oldPeer, newPeer *wgtypes.Peer) error {
	var modified PeerModifier = PeerModifiedNone

	// Compare peer properties
	if util.CmpEndpoint(newPeer.Endpoint, oldPeer.Endpoint) != 0 {
		modified |= PeerModifiedEndpoint
	}
	if newPeer.ProtocolVersion != oldPeer.ProtocolVersion {
		modified |= PeerModifiedProtocolVersion
	}
	if newPeer.PersistentKeepaliveInterval != oldPeer.PersistentKeepaliveInterval {
		modified |= PeerModifiedKeepaliveInterval
	}
	if newPeer.LastHandshakeTime != oldPeer.LastHandshakeTime {
		modified |= PeerModifiedHandshakeTime
	}
	if len(newPeer.AllowedIPs) != len(oldPeer.AllowedIPs) {
		modified |= PeerModifiedAllowedIPs
	} else {
		for i := 0; i < len(oldPeer.AllowedIPs); i++ {
			if util.CmpNet(&oldPeer.AllowedIPs[i], &newPeer.AllowedIPs[i]) != 0 {
				modified |= PeerModifiedAllowedIPs
				break
			}
		}
	}

	// Find changes in AllowedIP list
	// sort.Slice(newPeer.AllowedIPs, lessNets(newPeer.AllowedIPs))
	// sort.Slice(oldPeer.AllowedIPs, lessNets(oldPeer.AllowedIPs))

	// for i, j := 0, 0; i < len(oldPeer.AllowedIPs) && j < len(newPeer.AllowedIPs); {
	// 	oldAllowedIP := &oldPeer.AllowedIPs[i]
	// 	newAllowedIP := &newPeer.AllowedIPs[j]

	// 	cmp := cmpNet(oldAllowedIP, newAllowedIP)
	// 	switch {
	// 	case cmp < 0: // deleted
	// 		d.onPeerAllowedIPDeleted(oldPeer)

	// 	case cmp > 0: // added
	// 		d.onPeerAllowedIPAdded(newPeer)

	// 	default: //
	// 		i++
	// 		j++
	// 	}
	// }

	if modified != PeerModifiedNone {
		i.logger.WithFields(log.Fields{
			"peer":     oldPeer.PublicKey,
			"modified": modified.String(),
		}).Info("Peer modified")

		i.onPeerModified(oldPeer, newPeer, modified)
	}

	return nil
}

func (i *BaseInterface) Sync(newDev *wgtypes.Device) error {
	// Compare device properties
	if newDev.Type != i.Type {
		i.logger.WithFields(log.Fields{
			"old": i.Type,
			"new": newDev.Type,
		}).Info("Type changed")

		i.Device.Type = newDev.Type
	}
	if newDev.FirewallMark != i.FirewallMark {
		i.logger.WithFields(log.Fields{
			"old": i.FirewallMark,
			"new": newDev.FirewallMark,
		}).Info("FirewallMark changed")

		i.Device.FirewallMark = newDev.FirewallMark
	}
	if newDev.PrivateKey != i.Device.PrivateKey {
		i.logger.WithFields(log.Fields{
			"old": i.PrivateKey,
			"new": newDev.PrivateKey,
		}).Info("PrivateKey changed")

		i.Device.PrivateKey = newDev.PrivateKey
		i.Device.PublicKey = newDev.PublicKey
	}
	if newDev.ListenPort != i.ListenPort {
		i.logger.WithFields(log.Fields{
			"old": i.ListenPort,
			"new": newDev.ListenPort,
		}).Info("ListenPort changed")

		i.Device.ListenPort = newDev.ListenPort
	}

	sort.Slice(newDev.Peers, wg.LessPeers(newDev.Peers))
	sort.Slice(i.Device.Peers, wg.LessPeers(i.Peers))

	k, j := 0, 0
	for k < len(i.Peers) && j < len(newDev.Peers) {
		oldPeer := &i.Peers[k]
		newPeer := &newDev.Peers[j]

		cmp := wg.CmpPeers(oldPeer, newPeer)
		switch {
		case cmp < 0: // removed
			i.logger.WithField("peer", oldPeer.PublicKey).Info("Peer removed")
			i.onPeerRemoved(oldPeer)
			k++

		case cmp > 0: // added
			i.logger.WithField("peer", oldPeer.PublicKey).Info("Peer added")
			i.onPeerAdded(newPeer)
			j++

		default: //
			i.syncPeer(oldPeer, newPeer)
			k++
			j++
		}
	}

	for k < len(i.Peers) {
		oldPeer := &i.Peers[k]
		i.logger.WithField("peer", oldPeer.PublicKey).Info("Peer removed")
		i.onPeerRemoved(oldPeer)
		k++
	}

	for j < len(newDev.Peers) {
		newPeer := &newDev.Peers[j]
		i.logger.WithField("peer", newPeer.PublicKey).Info("Peer added")
		i.onPeerAdded(newPeer)
		j++
	}

	i.Peers = newDev.Peers
	i.lastSync = time.Now()

	return nil
}

func (i *BaseInterface) SyncConfig(cfg string) error {
	_, err := os.Stat(cfg)
	if err != nil {
		return err
	}

	// TODO: can we sync the config fully in Go?
	cmd := exec.Command("wg", "syncconf", i.Name(), cfg)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to sync configuration: %w\n%s", err, output)
	}

	i.logger.WithField("config", cfg).Debug("Synced configuration")

	return nil
}

func (i *BaseInterface) onPeerAdded(p *wgtypes.Peer) {
	peer, err := NewPeer(p, i)
	if err != nil {
		i.logger.WithError(err).WithField("peer", peer.PublicKey).Fatal("Failed to setup peer")
	}

	i.peers[peer.PublicKey()] = peer

	i.server.BroadcastEvent(&pb.Event{
		Type:  "peer",
		State: "added",
		Event: &pb.Event_Intf{
			Intf: &pb.InterfaceEvent{
				Interface: &pb.Interface{
					Name: i.Name(),
					Peers: []*pb.Peer{
						{
							PublicKey: peer.PublicKey().Bytes(),
						},
					},
				},
			},
		},
	})
}

func (i *BaseInterface) onPeerRemoved(p *wgtypes.Peer) {
	peer, ok := i.peers[crypto.Key(p.PublicKey)]
	if !ok {
		i.logger.WithField("peer", peer.PublicKey).Warn("Failed to find matching peer")
	}

	if err := peer.Close(); err != nil {
		i.logger.WithField("peer", peer.PublicKey).Warn("Failed to close peer")
	}

	i.server.BroadcastEvent(&pb.Event{
		Type:  "peer",
		State: "removed",
		Event: &pb.Event_Intf{
			Intf: &pb.InterfaceEvent{
				Interface: &pb.Interface{
					Name: i.Name(),
					Peers: []*pb.Peer{
						{
							PublicKey: peer.PublicKey().Bytes(),
						},
					},
				},
			},
		},
	})

	delete(i.peers, peer.PublicKey())
}

func (i *BaseInterface) onPeerModified(old, new *wgtypes.Peer, modified PeerModifier) {
	peer, ok := i.peers[crypto.Key(new.PublicKey)]
	if ok {
		peer.OnModified(new, modified)
	} else {
		i.logger.Error("Failed to find modified peer")
	}

	i.server.BroadcastEvent(&pb.Event{
		Type:  "peer",
		State: "modified",
		Event: &pb.Event_Intf{
			Intf: &pb.InterfaceEvent{
				Interface: &pb.Interface{
					Name: i.Name(),
					Peers: []*pb.Peer{
						{
							PublicKey: peer.PublicKey().Bytes(),
						},
					},
				},
			},
		},
	})
}

func (i *BaseInterface) AddPeer(pk wgtypes.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: pk,
			},
		},
	}

	return i.client.ConfigureDevice(i.Name(), cfg)
}

func (i *BaseInterface) RemovePeer(pk wgtypes.Key) error {
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{
			{
				PublicKey: pk,
				Remove:    true,
			},
		},
	}

	return i.client.ConfigureDevice(i.Name(), cfg)
}

func NewInterface(dev *wgtypes.Device, client *wgctrl.Client, backend signaling.Backend, server *socket.Server, args *args.Args) (BaseInterface, error) {
	i := BaseInterface{
		Device:  *dev,
		client:  client,
		backend: backend,
		server:  server,
		args:    args,
		logger: log.WithFields(log.Fields{
			"intf": dev.Name,
			"type": "kernel",
		}),
		peers: make(map[crypto.Key]Peer),
	}

	i.logger.Info("Creating new interface")

	// Sync config
	if i.args.ConfigSync {
		cfg := fmt.Sprintf("%s/%s.conf", i.args.ConfigPath, i.Name())
		if err := i.SyncConfig(cfg); err != nil {
			return BaseInterface{}, fmt.Errorf("failed to sync interface configuration: %w", err)
		}
	}

	// Fixup device config
	if err := i.Fixup(); err != nil {
		return BaseInterface{}, fmt.Errorf("failed to fix interface configuration: %w", err)
	}

	i.server.BroadcastEvent(&pb.Event{
		Type:  "interface",
		State: "added",
		Event: &pb.Event_Intf{
			Intf: &pb.InterfaceEvent{
				Interface: &pb.Interface{
					Name: i.Name(),
				},
			},
		},
	})

	// We remove all peers here so that they get added by the following sync
	i.Peers = nil
	i.Sync(dev)

	return i, nil
}

func (i *BaseInterface) Fixup() error {
	var cfg wgtypes.Config

	if !i.PrivateKey().IsSet() {
		if i.Type != wgtypes.Userspace {
			i.logger.Warn("Device has no private key. Generating one..")
		}

		key, _ := wgtypes.GeneratePrivateKey()

		cfg.PrivateKey = &key
	}

	if i.ListenPort == 0 {
		i.logger.Warn("Device has no listen port. Setting a random one..")

		// Ephemeral Port Range (RFC6056 Sect. 2.1)
		portMin := (1 << 15) + (1 << 14)
		portMax := (1 << 16)
		port := portMin + rand.Intn(portMax-portMin)

		cfg.ListenPort = &port
	}

	if err := i.client.ConfigureDevice(i.Name(), cfg); err != nil {
		return fmt.Errorf("failed to configure device: %w", err)
	}

	return nil
}
