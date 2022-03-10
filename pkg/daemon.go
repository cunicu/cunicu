package pkg

import (
	"fmt"
	"sync"
	"syscall"
	"time"

	"github.com/cilium/ebpf/rlimit"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/intf"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"

	"go.uber.org/zap/zapio"
)

type Daemon struct {
	Backend signaling.Backend
	Client  *wgctrl.Client
	Config  *config.Config

	Interfaces    intf.InterfaceList
	InterfaceLock sync.RWMutex

	Events chan *pb.Event

	eventListeners     map[chan *pb.Event]interface{}
	eventListenersLock sync.Mutex

	stopSig chan interface{}

	logger *zap.Logger
}

func NewDaemon(cfg *config.Config) (*Daemon, error) {
	var err error

	logger := zap.L().Named("daemon")
	events := make(chan *pb.Event, 16)

	// Create backend
	var backend signaling.Backend
	var community = crypto.GenerateKeyFromPassword(cfg.GetString("community"))

	if len(cfg.Backends) == 1 {
		backend, err = signaling.NewBackend(&signaling.BackendConfig{
			URI:       cfg.Backends[0],
			Community: &community,
		}, events)
	} else {
		backend, err = signaling.NewMultiBackend(cfg.Backends, &signaling.BackendConfig{
			Community: &community,
		}, events)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signaling backend: %w", err)
	}

	// Disable memlock for loading eBPF programs
	if err := rlimit.RemoveMemlock(); err != nil {
		panic(fmt.Errorf("failed to remove memlock: %w", err))
	}

	// Create Wireguard netlink socket
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create Wireguard client: %w", err)
	}

	d := &Daemon{
		Config:  cfg,
		Client:  client,
		Backend: backend,

		Interfaces:    intf.InterfaceList{},
		InterfaceLock: sync.RWMutex{},

		Events:         events,
		eventListeners: map[chan *pb.Event]interface{}{},

		stopSig: make(chan interface{}),

		logger: logger,
	}

	// Check if Wireguard interface can be created by the kernel
	if !cfg.IsSet("wg.userspace") {
		cfg.Set("wg.userspace", !intf.WireguardModuleExists())
	}

	return d, nil
}

func (d *Daemon) Run() error {
	ifEvents := make(chan intf.InterfaceEvent, 16)
	errors := make(chan error, 16)
	signals := internal.SetupSignals()

	if err := d.CreateInterfacesFromArgs(); err != nil {
		return fmt.Errorf("failed to create interfaces: %w", err)
	}

	if err := intf.WatchWireguardUserspaceInterfaces(ifEvents, errors); err != nil {
		return fmt.Errorf("failed to watch userspace interfaces: %w", err)
	}

	if err := intf.WatchWireguardKernelInterfaces(ifEvents, errors); err != nil {
		return fmt.Errorf("failed to watch kernel interfaces: %w", err)
	}

	d.logger.Debug("Starting initial interface sync")
	d.SyncAllInterfaces()

	ticker := time.NewTicker(d.Config.GetDuration("watch_interval"))

out:
	for {
		select {
		// We still a need periodic sync we can not (yet) monitor Wireguard interfaces
		// for changes via a netlink socket (patch is pending)
		case <-ticker.C:
			d.logger.Debug("Starting periodic interface sync")
			d.SyncAllInterfaces()

		case <-d.stopSig:
			d.logger.Info("Received stop request")
			break out

		case event := <-d.Events:
			if event.Time == nil {
				event.Time = pb.TimeNow()
			}

			d.eventListenersLock.Lock()
			for ch := range d.eventListeners {
				ch <- event
			}
			d.eventListenersLock.Unlock()

			event.Log(d.logger, "Broadcasted event", zap.Int("listeners", len(d.eventListeners)))

		case event := <-ifEvents:
			d.logger.Debug("Received interface event", zap.String("event", event.String()))
			d.SyncAllInterfaces()

		case err := <-errors:
			d.logger.Error("Failed to watch for interface changes", zap.Error(err))

		case sig := <-signals:
			d.logger.Debug("Received signal", zap.String("signal", sig.String()))
			switch sig {
			case syscall.SIGUSR1:
				d.SyncAllInterfaces()
			default:
				break out
			}
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

func (d *Daemon) GetInterfaceByName(name string) intf.Interface {
	for _, intf := range d.Interfaces {
		if intf.Name() == name {
			return intf
		}
	}

	return nil
}

func (d *Daemon) SyncAllInterfaces() error {
	devices, err := d.Client.Devices()
	if err != nil {
		d.logger.Fatal("Failed to list Wireguard interfaces", zap.Error(err))
	}

	syncedInterfaces := intf.InterfaceList{}
	keepInterfaces := intf.InterfaceList{}

	for _, device := range devices {
		if !d.Config.WireguardInterfaceFilter.Match([]byte(device.Name)) {
			continue // Skip interfaces which dont match the filter
		}

		// Find matching interface
		interf := d.GetInterfaceByName(device.Name)
		if interf == nil { // new interface
			d.logger.Info("Adding new interface", zap.String("intf", device.Name))

			i, err := intf.NewInterface(device, d.Client, d.Backend, d.Events, d.Config)
			if err != nil {
				d.logger.Fatal("Failed to create new interface",
					zap.Error(err),
					zap.String("intf", device.Name),
				)
			}

			interf = &i

			d.Interfaces = append(d.Interfaces, &i)
		} else { // existing interface
			d.logger.Debug("Sync existing interface", zap.String("intf", device.Name))

			if err := interf.Sync(device); err != nil {
				d.logger.Fatal("Failed to sync interface",
					zap.Error(err),
					zap.String("intf", device.Name),
				)
			}
		}

		syncedInterfaces = append(syncedInterfaces, interf)
	}

	for _, intf := range d.Interfaces {
		i := syncedInterfaces.GetByName(intf.Name())
		if i == nil {
			d.logger.Info("Removing vanished interface", zap.String("intf", intf.Name()))

			if err := intf.Close(); err != nil {
				d.logger.Fatal("Failed to close interface", zap.Error(err))
			}

			d.Events <- &pb.Event{
				Type:      pb.Event_INTERFACE_REMOVED,
				Interface: intf.Name(),
			}
		} else {
			keepInterfaces = append(keepInterfaces, intf)
		}
	}

	d.Interfaces = keepInterfaces

	return nil
}

func (d *Daemon) CreateInterfacesFromArgs() error {
	var devs intf.Devices
	devs, err := d.Client.Devices()
	if err != nil {
		return err
	}

	for _, interfName := range d.Config.WireguardInterfaces {
		dev := devs.GetByName(interfName)
		if dev != nil {
			d.logger.Warn("Interface already exists. Skipping..", zap.Any("intf", interfName))
			continue
		}

		var interf intf.Interface
		if d.Config.GetBool("wg.userspace") {
			interf, err = intf.CreateUserInterface(interfName, d.Client, d.Backend, d.Events, d.Config)
		} else {
			interf, err = intf.CreateKernelInterface(interfName, d.Client, d.Backend, d.Events, d.Config)
		}
		if err != nil {
			return fmt.Errorf("failed to create Wireguard device: %w", err)
		}

		if d.logger.Core().Enabled(zap.DebugLevel) {
			d.logger.Debug("Intialized interface:")
			interf.DumpConfig(&zapio.Writer{Log: d.logger})
		}

		d.Interfaces = append(d.Interfaces, interf)
	}

	return nil
}

func (d *Daemon) Stop() error {
	close(d.stopSig)

	return nil
}

func (d *Daemon) ListenEvents() chan *pb.Event {
	events := make(chan *pb.Event, 100)

	d.eventListenersLock.Lock()
	d.eventListeners[events] = nil
	d.eventListenersLock.Unlock()

	return events
}
