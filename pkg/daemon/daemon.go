// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/log"
	osx "github.com/stv0g/cunicu/pkg/os"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/wg"
)

var (
	errInsufficientPrivileges = errors.New("insufficient privileges. Please run cunicu with administrator privileges")
	ErrFeatureDeactivated     = errors.New("feature deactivated")
)

type Daemon struct {
	*Watcher

	Backend *signaling.MultiBackend
	Client  *wgctrl.Client
	Config  *config.Config

	devices []device.Device

	stop          chan any
	reexecOnClose bool

	logger *log.Logger
}

func NewDaemon(cfg *config.Config) (*Daemon, error) {
	var err error

	// Check permissions
	if !osx.HasAdminPrivileges() {
		return nil, errInsufficientPrivileges
	}

	d := &Daemon{
		Config:  cfg,
		devices: []device.Device{},
		stop:    make(chan any),
		logger:  log.Global.Named("daemon"),
	}

	// Create WireGuard netlink socket
	if d.Client, err = wgctrl.New(); err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %w", err)
	}

	// Create watcher
	if d.Watcher, err = NewWatcher(d.Client, cfg.WatchInterval, cfg.InterfaceFilter); err != nil {
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

	d.AddInterfaceHandler(d)

	return d, nil
}

// Start starts the daemon and blocks until Stop() is called.
func (d *Daemon) Start() error {
	if err := wg.CleanupUserSockets(); err != nil {
		return fmt.Errorf("failed to cleanup stale user space sockets: %w", err)
	}

	if err := d.CreateDevices(); err != nil {
		return fmt.Errorf("failed to create devices: %w", err)
	}

	go d.Watcher.Watch()

	if err := d.Watcher.Sync(); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

	signals := osx.SetupSignals(osx.SigUpdate)

out:
	for {
		select {
		case sig := <-signals:
			d.logger.Debug("Received signal", zap.Any("signal", sig))
			switch sig {
			case osx.SigUpdate:
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

// Stop stops the daemon
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
		if err := i.SyncFeatures(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Daemon) Close() error {
	if err := d.Watcher.Close(); err != nil {
		return fmt.Errorf("failed to close watcher: %w", err)
	}

	for _, i := range d.interfaces {
		if err := i.Close(); err != nil {
			return fmt.Errorf("failed to close interface: %w", err)
		}
	}

	if err := d.Backend.Close(); err != nil {
		return fmt.Errorf("failed to close signaling backend: %w", err)
	}

	for _, dev := range d.devices {
		if err := dev.Close(); err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}
	}

	if err := d.Client.Close(); err != nil {
		return fmt.Errorf("failed to close WireGuard client: %w", err)
	}

	if d.reexecOnClose {
		d.logger.Debug("Restarting daemon")
		return osx.ReexecSelf()
	}

	d.logger.Debug("Closed daemon")

	return nil
}

func (d *Daemon) CreateDevices() error {
	devs, err := d.Client.Devices()
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
