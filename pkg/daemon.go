package wice

import (
	"errors"
	"fmt"
	"net/url"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/feat"
	ac "riasc.eu/wice/pkg/feat/auto"
	ep "riasc.eu/wice/pkg/feat/disc/epice"
	cs "riasc.eu/wice/pkg/feat/sync/config"
	hs "riasc.eu/wice/pkg/feat/sync/hosts"
	rs "riasc.eu/wice/pkg/feat/sync/routes"
	"riasc.eu/wice/pkg/util"
	"riasc.eu/wice/pkg/watcher"
	"riasc.eu/wice/pkg/wg"

	"riasc.eu/wice/pkg/signaling"
)

type Daemon struct {
	*watcher.Watcher

	Features []feat.Feature

	EPDisc *ep.EndpointDiscovery

	// Shared
	Backend *signaling.MultiBackend
	client  *wgctrl.Client
	config  *config.Config

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
		config:  cfg,
		devices: []device.Device{},
		stop:    make(chan any),
	}

	d.logger = zap.L().Named("daemon")

	// Create WireGuard netlink socket
	d.client, err = wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create WireGuard client: %w", err)
	}

	// Create watcher
	if d.Watcher, err = watcher.New(d.client, cfg.WatchInterval, &cfg.WireGuard.InterfaceFilter.Regexp); err != nil {
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

	if err := d.setupFeatures(); err != nil {
		return nil, err
	}

	for _, feat := range d.Features {
		if err := feat.Start(); err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (d *Daemon) setupFeatures() error {
	if d.config.AutoConfig.Enabled {
		d.Features = append(d.Features, ac.New(d.Watcher, d.config, d.client))
	}

	if d.config.ConfigSync.Enabled {
		d.Features = append(d.Features, cs.New(d.Watcher, d.client, d.config.ConfigSync.Path, d.config.ConfigSync.Watch, d.config.WireGuard.Userspace))
	}

	if d.config.RouteSync.Enabled {
		d.Features = append(d.Features, rs.New(d.Watcher, d.config.RouteSync.Table))
	}

	if d.config.HostSync.Enabled {
		d.Features = append(d.Features, hs.New(d.Watcher))
	}

	if d.config.EndpointDisc.Enabled {
		d.EPDisc = ep.New(d.Watcher, d.config, d.client, d.Backend)
		d.Features = append(d.Features, d.EPDisc)
	}

	return nil
}

func (d *Daemon) Run() error {
	if err := wg.CleanupUserSockets(); err != nil {
		return fmt.Errorf("failed to cleanup stale userspace sockets: %w", err)
	}

	if err := d.CreateDevicesFromArgs(); err != nil {
		return fmt.Errorf("failed to create devices: %w", err)
	}

	signals := util.SetupSignals(util.SigUpdate)

	go d.Watcher.Run()

out:
	for {
		select {
		case sig := <-signals:
			d.logger.Debug("Received signal", zap.String("signal", sig.String()))
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

func (d *Daemon) Close() error {
	if err := d.Watcher.Close(); err != nil {
		return fmt.Errorf("failed to close interface: %w", err)
	}

	for _, feat := range d.Features {
		if err := feat.Close(); err != nil {
			return err
		}
	}

	if err := d.client.Close(); err != nil {
		return fmt.Errorf("failed to close WireGuard client: %w", err)
	}

	for _, dev := range d.devices {
		if err := dev.Close(); err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}
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

	if devs, err = d.client.Devices(); err != nil {
		return fmt.Errorf("failed to get existing WireGuard devices: %w", err)
	}

	for _, devName := range d.config.WireGuard.Interfaces {
		if !d.config.WireGuard.InterfaceFilter.MatchString(devName) {
			return fmt.Errorf("device '%s' is not matched by WireGuard interface filter '%s'",
				devName, d.config.WireGuard.InterfaceFilter.String())
		}

		if wgdev := devs.GetByName(devName); wgdev != nil {
			return fmt.Errorf("device '%s' already exists", devName)
		}

		dev, err := device.NewDevice(devName, d.config.WireGuard.Userspace)
		if err != nil {
			return fmt.Errorf("failed to create WireGuard device: %w", err)
		}

		d.devices = append(d.devices, dev)
	}

	return nil
}
