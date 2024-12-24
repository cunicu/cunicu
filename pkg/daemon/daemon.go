// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package daemon

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"cunicu.li/cunicu/pkg/config"
	"cunicu.li/cunicu/pkg/device"
	"cunicu.li/cunicu/pkg/log"
	osx "cunicu.li/cunicu/pkg/os"
	"cunicu.li/cunicu/pkg/os/systemd"
	"cunicu.li/cunicu/pkg/signaling"
	"cunicu.li/cunicu/pkg/wg"
)

var (
	errInsufficientPrivileges = errors.New("insufficient privileges. Please run cunicu with administrator privileges")
	ErrFeatureDeactivated     = errors.New("feature deactivated")
)

type State string

const (
	StateStarted       = "started"
	StateInitializing  = "initializing"
	StateReady         = "ready"
	StateReloading     = "reloading"
	StateStopping      = "stoppping"
	StateSynchronizing = "syncing"
)

type Daemon struct {
	*Watcher

	Backend *signaling.MultiBackend
	Client  *wgctrl.Client
	Config  *config.Config

	devices []device.Device

	state         State
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
		state:   StateStarted,
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
	if d.Backend, err = signaling.NewMultiBackend(cfg.Backends, &signaling.BackendConfig{}); err != nil {
		return nil, fmt.Errorf("failed to initialize signaling backend: %w", err)
	}

	d.AddInterfaceHandler(d)

	return d, nil
}

// Start starts the daemon and blocks until Shutdown() is called.
func (d *Daemon) Start() error {
	if err := d.setState(StateInitializing); err != nil {
		return fmt.Errorf("failed transition state: %w", err)
	}

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

	signals := osx.SetupSignals(osx.SigUpdate, osx.SigReload)

	wdt, err := d.watchdogTicker()
	if err != nil && !errors.Is(err, errNotSupported) {
		return fmt.Errorf("failed to get watchdog interval: %w", err)
	}

	if err := d.setState(StateReady); err != nil {
		return fmt.Errorf("failed transition state: %w", err)
	}

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

			case osx.SigReload:
				if err := d.reload(); err != nil {
					return err
				}

			default:
				break out
			}

		case <-wdt:
			if err := d.notify(systemd.NotifyWatchdog); err != nil {
				return fmt.Errorf("failed to notify systemd watchdog: %w", err)
			}
			d.logger.DebugV(20, "Watchdog tick")

		case <-d.stop:
			break out
		}
	}

	return nil
}

// Shutdown stops the daemon.
func (d *Daemon) Shutdown(restart bool) {
	if d.stop == nil {
		return
	}

	close(d.stop)
	d.stop = nil

	if restart {
		d.reexecOnClose = true
		d.logger.Debug("Restarting daemon")
	} else {
		d.logger.Debug("Stopping daemon")
	}
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
	if err := d.setState(StateStopping); err != nil {
		return fmt.Errorf("failed transition state: %w", err)
	}

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

func (d *Daemon) watchdogTicker() (<-chan time.Time, error) {
	wdInterval, err := systemd.WatchdogEnabled(true)
	if err != nil {
		return nil, err
	} else if wdInterval == 0 {
		d.logger.DebugV(5, "Not started via systemd. Disabling watchdog")
		return nil, errNotSupported
	}

	return time.NewTicker(wdInterval / 2).C, nil
}

func (d *Daemon) reload() error {
	if err := d.setState(StateReloading); err != nil {
		return fmt.Errorf("failed transition state: %w", err)
	}

	if _, err := d.Config.ReloadAllSources(); err != nil {
		d.logger.Error("Failed to reload config", zap.Error(err))
	}

	if err := d.setState(StateReady); err != nil {
		return fmt.Errorf("failed transition state: %w", err)
	}

	return nil
}

func (d *Daemon) setState(s State) error {
	d.state = s

	d.logger.DebugV(5, "Daemon state changed", zap.String("state", string(s)))

	switch d.state {
	case StateStarted:
	case StateInitializing:
	case StateSynchronizing:

	case StateReady:
		if err := d.notify(systemd.NotifyReady); err != nil {
			return fmt.Errorf("failed to notify systemd: %w", err)
		}

	case StateReloading:
		if err := d.notify(systemd.NotifyReloading); err != nil {
			return fmt.Errorf("failed to notify systemd: %w", err)
		}

	case StateStopping:
		if err := d.notify(systemd.NotifyStopping); err != nil {
			return fmt.Errorf("failed to notify systemd: %w", err)
		}
	}

	return nil
}

func (d *Daemon) notify(notify string) error {
	notifyMessages := []string{notify}

	if notify == systemd.NotifyReloading {
		now, err := osx.GetClockMonotonic()
		if err != nil {
			return fmt.Errorf("failed to get monotonic clock: %w", err)
		}

		notifyMessages = append(notifyMessages,
			fmt.Sprintf("MONOTONIC_USEC=%d", now.UnixMicro()))

		d.logger.DebugV(5, "Notifying systemd", zap.Strings("message", notifyMessages))
	}

	if _, err := systemd.Notify(false, strings.Join(notifyMessages, "\n")); err != nil {
		return err
	}

	return nil
}
