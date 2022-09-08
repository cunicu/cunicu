package wice

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/device"
	"github.com/stv0g/cunicu/pkg/feat"
	"github.com/stv0g/cunicu/pkg/feat/epdisc"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/watcher"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"

	"github.com/stv0g/cunicu/pkg/signaling"
)

type Daemon struct {
	*watcher.Watcher

	Features []feat.Feature

	EndpointDiscovery *epdisc.EndpointDiscovery

	// Shared
	Backend *signaling.MultiBackend
	Client  *wgctrl.Client
	Config  *config.Config

	devices []device.Device

	stop          chan any
	reexecOnClose bool

	logger *zap.Logger
}

func NewDaemon(cfg *config.Config) (*Daemon, error) {
	var err error

	// Check permissions
	if !util.HasAdminPrivileges() {
		return nil, errors.New("insufficient privileges. Please run ɯice as root user or with NET_ADMIN capabilities")
	}

	d := &Daemon{
		Config:  cfg,
		devices: []device.Device{},
		stop:    make(chan any),
	}

	d.logger = zap.L().Named("daemon")

	// Create WireGuard netlink socket
	d.Client, err = wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %w", err)
	}

	// Create watcher
	if d.Watcher, err = watcher.New(d.Client, cfg.WatchInterval, cfg.WireGuard.InterfaceFilter.Regexp.MatchString); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	// Create backend
	urls := []*url.URL{}
	for _, u := range cfg.Backends {
		urls = append(urls, &u.URL)
	}

	d.Backend, err = signaling.NewMultiBackend(urls, &signaling.BackendConfig{
		OnReady: []signaling.BackendReadyHandler{},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize signaling backend: %w", err)
	}

	// Check if WireGuard interface can be created by the kernel
	if !cfg.WireGuard.Userspace {
		cfg.WireGuard.Userspace = !wg.KernelModuleExists()
	}

	d.Features, d.EndpointDiscovery = feat.NewFeatures(d.Watcher, d.Config, d.Client, d.Backend)

	for _, feat := range d.Features {
		if err := feat.Start(); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (d *Daemon) Run() error {
	if err := wg.CleanupUserSockets(); err != nil {
		return fmt.Errorf("failed to cleanup stale userspace sockets: %w", err)
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

	for _, f := range d.Features {
		if s, ok := f.(feat.Syncable); ok {
			if err := s.Sync(); err != nil {
				return err
			}
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
		return fmt.Errorf("failed to close interface: %w", err)
	}

	for _, feat := range d.Features {
		if err := feat.Close(); err != nil {
			return err
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
	var devs wg.Devices
	var err error

	if devs, err = d.Client.Devices(); err != nil {
		return fmt.Errorf("failed to get existing WireGuard devices: %w", err)
	}

	for _, devName := range d.Config.WireGuard.Interfaces {
		if !d.Config.WireGuard.InterfaceFilter.MatchString(devName) {
			return fmt.Errorf("device '%s' is not matched by WireGuard interface filter '%s'",
				devName, d.Config.WireGuard.InterfaceFilter.String())
		}

		if wgdev := devs.GetByName(devName); wgdev != nil {
			return fmt.Errorf("device '%s' already exists", devName)
		}

		dev, err := device.NewDevice(devName, d.Config.WireGuard.Userspace)
		if err != nil {
			return fmt.Errorf("failed to create WireGuard device: %w", err)
		}

		d.devices = append(d.devices, dev)
	}

	return nil
}
