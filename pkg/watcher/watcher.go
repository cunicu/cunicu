// Package watcher keeps track and monitors for new, removed and modified WireGuard interfaces and peers.
package watcher

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/util"
)

const (
	InterfaceAdded InterfaceEventOp = iota
	InterfaceDeleted
)

type InterfaceEventOp int
type InterfaceEvent struct {
	Op   InterfaceEventOp
	Name string
}

func (ls InterfaceEventOp) String() string {
	switch ls {
	case InterfaceAdded:
		return "added"
	case InterfaceDeleted:
		return "deleted"
	default:
		return ""
	}
}

func (e InterfaceEvent) String() string {
	return fmt.Sprintf("%s %s", e.Name, e.Op)
}

// Watcher monitors both userspace and kernel for changes to WireGuard interfaces
type Watcher struct {
	Interfaces    core.InterfaceList
	InterfaceLock sync.RWMutex

	devices []*wgtypes.Device

	onInterface []core.InterfaceHandler

	client *wgctrl.Client

	events chan InterfaceEvent
	errors chan error
	stop   chan any

	// Settings
	filter   *regexp.Regexp
	interval time.Duration

	logger *zap.Logger
}

func New(client *wgctrl.Client, interval time.Duration, filter *regexp.Regexp) (*Watcher, error) {
	return &Watcher{
		Interfaces:    core.InterfaceList{},
		InterfaceLock: sync.RWMutex{},

		onInterface: []core.InterfaceHandler{},

		devices: []*wgtypes.Device{},

		client:   client,
		filter:   filter,
		interval: interval,

		events: make(chan InterfaceEvent, 16),
		errors: make(chan error, 16),
		stop:   make(chan any),

		logger: zap.L().Named("watcher"),
	}, nil
}

func (w *Watcher) Close() error {
	if err := w.Sync(); err != nil {
		return fmt.Errorf("final sync failed")
	}

	close(w.stop)

	return nil
}

func (w *Watcher) Run() {
	w.logger.Debug("Started initial synchronization")
	if err := w.Sync(); err != nil {
		w.logger.Fatal("Initial synchronization failed", zap.Error(err))
	}
	w.logger.Debug("Finished initial synchronization")

	if err := w.watchUser(); err != nil {
		w.logger.Fatal("Failed to watch userspace interfaces", zap.Error(err))
	}
	w.logger.Debug("Started watching for changes of WireGuard userspace devices")

	if err := w.watchKernel(); err != nil {
		w.logger.Fatal("Failed to watch kernel interfaces", zap.Error(err))
	}
	w.logger.Debug("Started watching for changes of WireGuard kernel devices")

	ticker := time.NewTicker(w.interval)

out:
	for {
		select {
		// We still a need periodic sync we can not (yet) monitor WireGuard interfaces
		// for changes via a netlink socket (patch is pending)
		case <-ticker.C:
			w.logger.Debug("Started periodic interface synchronization")
			if err := w.Sync(); err != nil {
				w.logger.Error("Synchronization failed", zap.Error(err))
			}
			w.logger.Debug("Completed periodic interface synchronization")

		case event := <-w.events:
			w.logger.Debug("Received interface event", zap.String("event", event.String()))
			if err := w.Sync(); err != nil {
				w.logger.Error("Synchronization failed", zap.Error(err))
			}

		case err := <-w.errors:
			w.logger.Error("Failed to watch for interface changes", zap.Error(err))

		case <-w.stop:
			break out
		}
	}
}

func (w *Watcher) Sync() error {
	var err error

	var new = []*wgtypes.Device{}
	var old = w.devices

	if new, err = w.client.Devices(); err != nil {
		return fmt.Errorf("failed to list WireGuard interfaces: %w", err)
	}

	// Ignore devices which do not match the filter
	new = util.FilterSlice(new, func(d *wgtypes.Device) bool {
		return w.filter == nil || w.filter.Match([]byte(d.Name))
	})

	added, removed, kept := util.DiffSliceFunc(old, new, func(a, b **wgtypes.Device) int {
		return strings.Compare((*a).Name, (*b).Name)
	})

	for _, wgd := range removed {
		i, ok := w.Interfaces[wgd.Name]
		if !ok {
			w.logger.Warn("Failed to find matching interface", zap.Any("intf", wgd.Name))
			continue
		}

		w.logger.Info("Interface removed", zap.String("intf", wgd.Name))

		for _, h := range w.onInterface {
			h.OnInterfaceRemoved(i)
		}

		delete(w.Interfaces, wgd.Name)
	}

	for _, wgd := range added {
		w.logger.Info("Interface added", zap.String("intf", wgd.Name))

		i, err := core.NewInterface(wgd, w.client)
		if err != nil {
			w.logger.Fatal("Failed to create new interface",
				zap.Error(err),
				zap.String("intf", wgd.Name),
			)
		}

		// We purposefully prune the peer list here to force full initial sync of all peers
		i.Device.Peers = nil

		w.Interfaces[wgd.Name] = i

		for _, h := range w.onInterface {
			h.OnInterfaceAdded(i)
		}

		i.Sync(wgd)
	}

	for _, wgd := range kept {
		i, ok := w.Interfaces[wgd.Name]
		if !ok {
			w.logger.Warn("Failed to find matching interface", zap.Any("intf", wgd.Name))
			continue
		}

		i.Sync(wgd)
	}

	w.devices = new

	return nil
}

func (w *Watcher) Peer(intf string, pk *crypto.Key) *core.Peer {
	i := w.Interfaces.ByName(intf)
	if i == nil {
		return nil
	}

	if p, ok := i.Peers[*pk]; ok {
		return p
	}

	return nil
}

func (w *Watcher) PeerByKey(pk *crypto.Key) *core.Peer {
	for _, i := range w.Interfaces {
		if p, ok := i.Peers[*pk]; ok {
			return p
		}
	}

	return nil
}
