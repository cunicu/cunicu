package wice

import (
	"errors"
	"fmt"
	"net/url"

	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/device"
	errs "riasc.eu/wice/pkg/errors"
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

	// Features

	AutoConfig *ac.AutoConfig
	ConfigSync *cs.ConfigSync
	HostsSync  *hs.HostsSync
	RouteSync  *rs.RouteSync
	EPDisc     *ep.EndpointDiscovery

	// Shared

	Backend *signaling.MultiBackend
	client  *wgctrl.Client
	config  *config.Config

	devices []device.Device

	stop chan any

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

	return d, nil
}

func (d *Daemon) setupFeatures() error {
	var err error

	if d.config.AutoConfig.Enabled {
		if d.AutoConfig, err = ac.New(d.Watcher, d.config, d.client); err != nil {
			return fmt.Errorf("failed to start interface auto configuration: %w", err)
		}
	}

	if d.config.ConfigSync.Enabled {
		if d.ConfigSync, err = cs.New(d.Watcher, d.client,
			d.config.ConfigSync.Path,
			d.config.ConfigSync.Watch,
			d.config.WireGuard.Userspace); err != nil {

			return fmt.Errorf("failed to start configuration file synchronization: %w", err)
		}

		d.logger.Info("Started configuration file synchronization")
	}

	if d.config.RouteSync.Enabled {
		if d.RouteSync, err = rs.New(d.Watcher, d.config.RouteSync.Table); err != nil {
			return fmt.Errorf("failed to start allowed-ips <-> kernel route synchronization: %w", err)
		}

		d.logger.Info("Started route synchronization")
	}

	if d.config.HostSync.Enabled {
		if d.HostsSync, err = hs.New(d.Watcher); err != nil {
			return fmt.Errorf("failed to start host name synchronization: %w", err)
		}

		d.logger.Info("Started /etc/hosts synchronization")
	}

	if d.config.EndpointDisc.Enabled {
		if d.EPDisc, err = ep.New(d.Watcher, d.config, d.client, d.Backend); err != nil {
			return fmt.Errorf("failed to start endpoint discovery: %w", err)
		}

		d.logger.Info("Started ICE endpoint discovery")
	}

	return nil
}

func (d *Daemon) closeFeatures() error {
	if d.EPDisc != nil {
		if err := d.EPDisc.Close(); err != nil {
			return fmt.Errorf("failed to stop endpoint discovery: %w", err)
		}
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

func (d *Daemon) IsRunning() bool {
	select {
	case _, running := <-d.stop:
		return running
	default:
		return true
	}
}

func (d *Daemon) Stop() error {
	if !d.IsRunning() {
		return errs.ErrAlreadyStopped
	}

	close(d.stop)

	return nil
}

func (d *Daemon) Close() error {
	if err := d.Stop(); err != nil && !errors.Is(err, errs.ErrAlreadyStopped) {
		return err
	}

	if err := d.closeFeatures(); err != nil {
		return err
	}

	if err := d.Watcher.Close(); err != nil {
		return fmt.Errorf("failed to close interface: %w", err)
	}

	if err := d.client.Close(); err != nil {
		return fmt.Errorf("failed to close WireGuard client: %w", err)
	}

	for _, dev := range d.devices {
		if err := dev.Delete(); err != nil {
			return fmt.Errorf("failed to delete device: %w", err)
		}
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
