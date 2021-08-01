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

	"github.com/pion/ice/v2"
	log "github.com/sirupsen/logrus"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/internal/wg"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/crypto"
)

const (
	PeerModifiedEndpoint          = (1 << 0)
	PeerModifiedKeepaliveInterval = (1 << 1)
	PeerModifiedProtocolVersion   = (1 << 2)
	PeerModifiedAllowedIPs        = (1 << 3)
	PeerModifiedHandshakeTime     = (1 << 4)
)

type BaseInterface struct {
	wgtypes.Device

	peers map[crypto.Key]Peer

	lastSync time.Time

	logger *log.Entry

	client  *wgctrl.Client
	mux     ice.UDPMux
	backend backend.Backend
	args    *args.Args
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
		err := p.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *BaseInterface) DumpConfig(wr io.Writer) {
	fmt.Fprintf(wr, "[Interface] # %s\n", i.Name())
	if i.PublicKey().IsSet() {
		fmt.Fprintf(wr, "PublicKey = %s\n", i.PublicKey)
	}
	if i.PrivateKey().IsSet() {
		fmt.Fprintf(wr, "PrivateKey = %s\n", i.PrivateKey)
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
	modified := 0

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

	if modified != 0 {
		i.logger.WithField("peer", oldPeer.PublicKey).WithField("modified", modified).Info("Peer modified")
		i.onPeerModified(oldPeer, newPeer, modified)
	}

	return nil
}

func (i *BaseInterface) Sync(newDev *wgtypes.Device) error {
	// Compare device properties
	if newDev.Type != i.Type {
		i.logger.WithField("old", i.Type).WithField("new", newDev.Type).Info("Type changed")
		i.Device.Type = newDev.Type
	}
	if newDev.FirewallMark != i.FirewallMark {
		i.logger.WithField("old", i.FirewallMark).WithField("new", newDev.FirewallMark).Info("FirewallMark changed")
		i.Device.FirewallMark = newDev.FirewallMark
	}
	if newDev.PrivateKey != i.Device.PrivateKey {
		i.logger.WithField("old", i.PrivateKey).WithField("new", newDev.PrivateKey).Info("PrivateKey changed")
		i.Device.PrivateKey = newDev.PrivateKey
		i.Device.PublicKey = newDev.PublicKey
	}
	if newDev.ListenPort != i.ListenPort {
		i.logger.WithField("old", i.ListenPort).WithField("new", newDev.ListenPort).Info("ListenPort changed")
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
}

func (i *BaseInterface) onPeerRemoved(peer *wgtypes.Peer) {
	p, ok := i.peers[crypto.Key(peer.PublicKey)]
	if !ok {
		i.logger.WithField("peer", peer.PublicKey).Warn("Failed to find matching peer")
	}

	err := p.Close()
	if err != nil {
		i.logger.WithField("peer", peer.PublicKey).Warn("Failed to close peer")
	}

	delete(i.peers, crypto.Key(peer.PublicKey))
}

func (i *BaseInterface) onPeerModified(old, new *wgtypes.Peer, modified int) {
	p, ok := i.peers[crypto.Key(new.PublicKey)]
	if ok {
		p.OnModified(new, modified)
	} else {
		i.logger.Error("Failed to find modified peer")
	}
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

func NewInterface(dev *wgtypes.Device, client *wgctrl.Client, backend backend.Backend, args *args.Args) (BaseInterface, error) {
	i := BaseInterface{
		Device:  *dev,
		client:  client,
		backend: backend,
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
		err := i.SyncConfig(cfg)
		if err != nil {
			return BaseInterface{}, fmt.Errorf("failed to sync interface configuration: %w", err)
		}
	}

	// Fixup device config
	err := i.Fixup()
	if err != nil {
		return BaseInterface{}, fmt.Errorf("failed to fix interface configuration: %w", err)
	}

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

	err := i.client.ConfigureDevice(i.Name(), cfg)
	if err != nil {
		return fmt.Errorf("failed to configure device: %w", err)
	}

	return nil
}
