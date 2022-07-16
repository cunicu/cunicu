package pkg

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
	"golang.zx2c4.com/wireguard/wgctrl"
	"kernel.org/pub/linux/libs/security/libcap/cap"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/internal/util"
	"riasc.eu/wice/internal/wg"
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/device"
	"riasc.eu/wice/pkg/feat/disc/ice"
	"riasc.eu/wice/pkg/feat/setup"
	config_sync "riasc.eu/wice/pkg/feat/sync/config"
	route_sync "riasc.eu/wice/pkg/feat/sync/routes"
	"riasc.eu/wice/pkg/watcher"

	"riasc.eu/wice/pkg/signaling"

	"go.uber.org/zap/zapio"
)

type Daemon struct {
	*watcher.Watcher

	// Features

	ConfigSyncer      *config_sync.Syncer
	RouteSyncer       *route_sync.Syncer
	Setup             *setup.Setup
	EndpointDiscovery *ice.EndpointDiscovery

	// Shared

	Backend signaling.Backend
	client  *wgctrl.Client
	config  *config.Config

	stop    chan any
	signals chan os.Signal

	logger *zap.Logger
}

func NewDaemon(cfg *config.Config) (*Daemon, error) {
	var err error

	logger := zap.L().Named("daemon")

	// Check permissions
	if !util.HasCapabilities(cap.NET_ADMIN) {
		return nil, errors.New("insufficient privileges. Pleas run wice as root user or with NET_ADMIN capabilities")
	}

	// Create backend
	var backend signaling.Backend

	if len(cfg.Backends) == 1 {
		backend, err = signaling.NewBackend(&signaling.BackendConfig{
			URI: &cfg.Backends[0].URL,
		})
	} else {
		urls := []*url.URL{}
		for _, u := range cfg.Backends {
			urls = append(urls, &u.URL)
		}

		backend, err = signaling.NewMultiBackend(urls, &signaling.BackendConfig{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signaling backend: %w", err)
	}

	// Create Wireguard netlink socket
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Wireguard client: %w", err)
	}

	d := &Daemon{
		config:  cfg,
		client:  client,
		Backend: backend,

		stop:    make(chan any),
		signals: internal.SetupSignals(),

		logger: logger,
	}

	if d.Watcher, err = watcher.New(d.client, cfg.WatchInterval, &cfg.Wireguard.InterfaceFilter.Regexp); err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}

	// Check if Wireguard interface can be created by the kernel
	if !cfg.Wireguard.Userspace {
		cfg.Wireguard.Userspace = !wg.KernelModuleExists()
	}

	if err := d.setupFeatures(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Daemon) setupFeatures() error {
	var err error

	// TODO: Add configuration setting
	if true {
		if d.Setup, err = setup.New(d.Watcher, d.client); err != nil {
			return fmt.Errorf("failed to create interface setuper: %w", err)
		}
	}

	if d.config.Wireguard.Config.Sync {
		if d.ConfigSyncer, err = config_sync.New(d.Watcher, d.client,
			d.config.Wireguard.Config.Path,
			d.config.Wireguard.Config.Watch,
			d.config.Wireguard.Userspace); err != nil {

			return fmt.Errorf("failed to start configuration file synchronization: %w", err)
		}

		d.logger.Info("Started configuration file synchronization")
	}

	if d.config.Wireguard.Routes.Sync {
		if d.RouteSyncer, err = route_sync.New(d.Watcher, d.config.Wireguard.Routes.Table); err != nil {
			return fmt.Errorf("failed to start allowed-ips <-> kernel route synchronization: %w", err)
		}

		d.logger.Info("Started allowed-ips <-> kernel route synchronization")
	}

	// TODO: Add configuration setting
	if true {
		if d.EndpointDiscovery, err = ice.New(d.Watcher, d.config, d.client, d.Backend); err != nil {
			return fmt.Errorf("failed to start endpoint discovery: %w", err)
		}

		d.logger.Info("Started endpoint discovery")
	}

	return nil
}

func (d *Daemon) Run() error {
	if err := d.CreateInterfacesFromArgs(); err != nil {
		return fmt.Errorf("failed to create interfaces: %w", err)
	}

	go d.Watcher.Run()

	for sig := range d.signals {
		d.logger.Debug("Received signal", zap.String("signal", sig.String()))
		switch sig {
		case unix.SIGUSR1:
			if err := d.Sync(); err != nil {
				d.logger.Error("Failed to synchronize interfaces", zap.Error(err))
			}
		default:
			return nil
		}
	}

	return nil
}

func (d *Daemon) Close() error {
	if err := d.Interfaces.Close(); err != nil {
		return fmt.Errorf("failed to close interface: %w", err)
	}

	return nil
}

func (d *Daemon) CreateInterfacesFromArgs() error {
	var devs device.Devices
	devs, err := d.client.Devices()
	if err != nil {
		return err
	}

	for _, intfName := range d.config.Wireguard.Interfaces {
		dev := devs.GetByName(intfName)
		if dev != nil {
			d.logger.Warn("Interface already exists. Skipping..", zap.Any("intf", intfName))
			continue
		}

		i, err := core.CreateInterface(intfName, d.config.Wireguard.Userspace, d.client)
		if err != nil {
			return fmt.Errorf("failed to create Wireguard device: %w", err)
		}

		if d.logger.Core().Enabled(zap.DebugLevel) {
			d.logger.Debug("Initialized interface:")
			if err := i.DumpConfig(&zapio.Writer{Log: d.logger}); err != nil {
				return err
			}
		}

		d.Watcher.Interfaces[i.Name()] = i
	}

	return nil
}

func (d *Daemon) Stop() error {
	close(d.stop)

	return nil
}
