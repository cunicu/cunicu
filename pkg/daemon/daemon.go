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
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/signaling"
)

type Daemon struct {
	*watcher.Watcher

	// Shared
	Backend *signaling.MultiBackend
	Client  *wgctrl.Client
	Config  *config.Config

	devices    []device.Device
	interfaces map[*core.Interface]*Interface

	stop          chan any
	reexecOnClose bool

	logger *zap.Logger
}

func New(cfg *config.Config) (*Daemon, error) {
	var err error

	// Check permissions
	if !util.HasAdminPrivileges() {
		return nil, errors.New("insufficient privileges. Please run cunicu as root user or with NET_ADMIN capabilities")
	}

	d := &Daemon{
		Config:     cfg,
		devices:    []device.Device{},
		interfaces: map[*core.Interface]*Interface{},
		stop:       make(chan any),
	}

	d.logger = zap.L().Named("daemon")

	// Initialize some defaults configuration settings at runtime
	if err = config.InitDefaults(); err != nil {
		return nil, fmt.Errorf("failed to initialize defaults: %w", err)
	}

	// Create WireGuard netlink socket
	d.Client, err = wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %w", err)
	}

	// Create watcher
	if d.Watcher, err = watcher.New(d.Client, cfg.WatchInterval, cfg.InterfaceFilter); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	// Create backend
	urls := []*url.URL{}
	for _, u := range cfg.Backends {
		urls = append(urls, &u.URL)
	}

	if d.Backend, err = signaling.NewMultiBackend(urls, &signaling.BackendConfig{
		OnReady: []signaling.BackendReadyHandler{},
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize signaling backend: %w", err)
	}

	d.Watcher.OnInterface(d)

	return d, nil
}

func (d *Daemon) Run() error {
	if err := wg.CleanupUserSockets(); err != nil {
		return fmt.Errorf("failed to cleanup stale user space sockets: %w", err)
	}

	if err := d.CreateDevicesFromArgs(); err != nil {
		return fmt.Errorf("failed to create devices: %w", err)
	}

	signals := util.SetupSignals(util.SigUpdate)

	d.logger.Debug("Started initial synchronization")
	if err := d.Watcher.Sync(); err != nil {
		d.logger.Fatal("Initial synchronization failed", zap.Error(err))
	}
	d.logger.Debug("Finished initial synchronization")

	go d.Watcher.Watch()

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
	if err := d.Watcher.Sync(); err != nil {
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
	for _, dev := range d.devices {
		if err := dev.Close(); err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}
	}

	if err := d.Watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	for _, i := range d.interfaces {
		if err := i.Close(); err != nil {
			return fmt.Errorf("failed to close interface: %w", err)
		}
	}

	if err := d.Client.Close(); err != nil {
		return fmt.Errorf("failed to close WireGuard client: %w", err)
	}

	d.logger.Debug("Closed daemon")

	if d.reexecOnClose {
		return util.ReexecSelf()
	}

	return nil
}

func (d *Daemon) CreateDevicesFromArgs() error {
	var devs wg.DeviceList
	var err error

	if devs, err = d.Client.Devices(); err != nil {
		return fmt.Errorf("failed to get existing WireGuard devices: %w", err)
	}

	isPattern := func(s string) bool {
		return strings.ContainsAny(s, "*?[]\\")
	}

	for name := range d.Config.Interfaces {
		if isPattern(name) {
			continue
		}

		if wgdev := devs.GetByName(name); wgdev != nil {
			// Device already exists
			continue
		}

		icfg := d.Config.InterfaceSettings(name)

		dev, err := device.NewDevice(name, icfg.WireGuard.UserSpace)
		if err != nil {
			return fmt.Errorf("failed to create WireGuard device: %w", err)
		}

		d.devices = append(d.devices, dev)
	}

	return nil
}

func (d *Daemon) OnInterfaceAdded(ci *core.Interface) {
	i, err := d.NewInterface(ci)
	if err != nil {
		d.logger.Error("Failed to add interface", zap.Error(err))
	}

	d.interfaces[ci] = i

	if err := i.Start(); err != nil {
		d.logger.Error("Failed to start interface", zap.Error(err))
	}
}

func (d *Daemon) OnInterfaceRemoved(ci *core.Interface) {
	i := d.interfaces[ci]

	if err := i.Close(); err != nil {
		d.logger.Error("Failed to close interface", zap.Error(err))
	}

	delete(d.interfaces, ci)
}

// Simple wrappers for d.Watcher.InterfaceBy*

func (d *Daemon) InterfaceByCore(ci *core.Interface) *Interface {
	return d.interfaces[ci]
}

func (d *Daemon) InterfaceByName(name string) *Interface {
	ci := d.Watcher.InterfaceByName(name)
	if ci == nil {
		return nil
	}

	return d.interfaces[ci]
}

func (d *Daemon) InterfaceByPublicKey(pk crypto.Key) *Interface {
	ci := d.Watcher.InterfaceByPublicKey(pk)
	if ci == nil {
		return nil
	}

	return d.interfaces[ci]
}

func (d *Daemon) InterfaceByIndex(idx int) *Interface {
	ci := d.Watcher.InterfaceByIndex(idx)
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
