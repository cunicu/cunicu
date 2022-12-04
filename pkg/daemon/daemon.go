package daemon

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
)

var errInsufficientPrivileges = errors.New("insufficient privileges. Please run cunicu with administrator privileges")

type Daemon struct {
	// Shared
	Backend *signaling.MultiBackend
	client  *wgctrl.Client
	Config  *config.Config

	watcher    *watcher.Watcher
	devices    []device.Device
	interfaces map[*core.Interface]*Interface

	onInterface []InterfaceHandler

	stop          chan any
	reexecOnClose bool

	logger *zap.Logger
}

func New(cfg *config.Config) (*Daemon, error) {
	var err error

	// Check permissions
	if !util.HasAdminPrivileges() {
		return nil, errInsufficientPrivileges
	}

	d := &Daemon{
		Config:      cfg,
		devices:     []device.Device{},
		interfaces:  map[*core.Interface]*Interface{},
		onInterface: []InterfaceHandler{},
		stop:        make(chan any),
	}

	d.logger = zap.L().Named("daemon")

	// Create WireGuard netlink socket
	if d.client, err = wgctrl.New(); err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %w", err)
	}

	// Create watcher
	if d.watcher, err = watcher.New(d.client, cfg.WatchInterval, cfg.InterfaceFilter); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	// Create signaling backend
	urls := []*url.URL{}
	for _, u := range cfg.Backends {
		u := u
		urls = append(urls, &u.URL)
	}

	if d.Backend, err = signaling.NewMultiBackend(urls, &signaling.BackendConfig{}); err != nil {
		return nil, fmt.Errorf("failed to initialize signaling backend: %w", err)
	}

	d.watcher.OnInterface(d)

	return d, nil
}

func (d *Daemon) Run() error {
	if err := wg.CleanupUserSockets(); err != nil {
		return fmt.Errorf("failed to cleanup stale user space sockets: %w", err)
	}

	if err := d.CreateDevices(); err != nil {
		return fmt.Errorf("failed to create devices: %w", err)
	}

	signals := util.SetupSignals(util.SigUpdate)

	go d.watcher.Watch()

	if err := d.watcher.Sync(); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

out:
	for {
		select {
		case sig := <-signals:
			d.logger.Debug("Received signal", zap.Any("signal", sig))
			switch sig {
			case util.SigUpdate:
				if err := d.Sync(); err != nil {
					d.logger.Error("Failed to synchronize interfaces", zap.Error(err))
				}
			default:
				break out
			}

		case <-d.stop:
			break out
		}
	}

	return nil
}

func (d *Daemon) Stop() {
	close(d.stop)
	d.logger.Debug("Stopping daemon")
}

func (d *Daemon) Restart() {
	d.reexecOnClose = true
	close(d.stop)
	d.logger.Debug("Restarting daemon")
}

func (d *Daemon) Sync() error {
	if err := d.watcher.Sync(); err != nil {
		return err
	}

	for _, i := range d.interfaces {
		if err := i.Sync(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Daemon) Close() error {
	if err := d.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	for _, i := range d.interfaces {
		if err := i.Close(); err != nil {
			return fmt.Errorf("failed to close interface: %w", err)
		}
	}

	for _, dev := range d.devices {
		if err := dev.Close(); err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}
	}

	if err := d.client.Close(); err != nil {
		return fmt.Errorf("failed to close WireGuard client: %w", err)
	}

	d.logger.Debug("Closed daemon")

	if d.reexecOnClose {
		return util.ReexecSelf()
	}

	return nil
}

func (d *Daemon) CreateDevices() error {
	devs, err := d.client.Devices()
	if err != nil {
		return fmt.Errorf("failed to get existing WireGuard devices: %w", err)
	}

	isPattern := func(s string) bool {
		return strings.ContainsAny(s, "*?[]\\")
	}

	alreadyExists := func(s string) bool {
		for _, dev := range devs {
			if dev.Name == s {
				return true
			}
		}

		for _, dev := range d.devices {
			if dev.Name() == s {
				return true
			}
		}

		return false
	}

	for _, name := range d.Config.InterfaceOrder {
		if isPattern(name) {
			continue // Patterns are ignored
		}

		if alreadyExists(name) {
			continue // Device already exists
		}

		icfg := d.Config.InterfaceSettings(name)

		dev, err := device.NewDevice(name, icfg.UserSpace)
		if err != nil {
			return fmt.Errorf("failed to create WireGuard device: %w", err)
		}

		d.devices = append(d.devices, dev)
	}

	return nil
}

// Simple wrappers for d.Watcher.InterfaceBy*

func (d *Daemon) InterfaceByCore(ci *core.Interface) *Interface {
	return d.interfaces[ci]
}

func (d *Daemon) InterfaceByName(name string) *Interface {
	ci := d.watcher.InterfaceByName(name)
	if ci == nil {
		return nil
	}

	return d.interfaces[ci]
}

func (d *Daemon) InterfaceByPublicKey(pk crypto.Key) *Interface {
	ci := d.watcher.InterfaceByPublicKey(pk)
	if ci == nil {
		return nil
	}

	return d.interfaces[ci]
}

func (d *Daemon) InterfaceByIndex(idx int) *Interface {
	ci := d.watcher.InterfaceByIndex(idx)
	if ci == nil {
		return nil
	}

	return d.interfaces[ci]
}

func (d *Daemon) ForEachInterface(cb func(i *Interface) error) error {
	for _, i := range d.interfaces {
		if err := cb(i); err != nil {
			return err
		}
	}

	return nil
}
