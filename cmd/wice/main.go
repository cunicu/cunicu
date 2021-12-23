package main

import (
	"os"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"golang.zx2c4.com/wireguard/wgctrl"

	"riasc.eu/wice/internal"
	"riasc.eu/wice/pkg/args"
	"riasc.eu/wice/pkg/intf"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"

	_ "riasc.eu/wice/pkg/signaling/k8s"
	_ "riasc.eu/wice/pkg/signaling/p2p"
)

func main() {
	internal.SetupRand()
	signals := internal.SetupSignals()
	logger := internal.SetupLogging()
	defer logger.Sync()

	args, err := args.Parse(os.Args[0], os.Args[1:])
	if err != nil {
		logger.Fatal("Failed to parse arguments", zap.Error(err))
	}

	if logger.Core().Enabled(zap.DebugLevel) {
		wr := &zapio.Writer{Log: logger}
		args.DumpConfig(wr)
	}

	// Create control socket server
	server, err := socket.Listen("unix", args.Socket, args.SocketWait)
	if err != nil {
		logger.Fatal("Failed to initialize control socket", zap.Error(err))
	}

	// Create backend
	var backend signaling.Backend
	if len(args.Backends) == 1 {
		backend, err = signaling.NewBackend(args.Backends[0], server)
	} else {
		backend, err = signaling.NewMultiBackend(args.Backends, server)
	}
	if err != nil {
		logger.Fatal("Failed to initialize backend", zap.Error(err))
	}

	// Create Wireguard netlink socket
	client, err := wgctrl.New()
	if err != nil {
		logger.Fatal("Failed to create Wireguard client", zap.Error(err))
	}

	// Create interfaces
	interfaces := &intf.Interfaces{}
	defer interfaces.CloseAll()

	interfaces.CreateFromArgs(client, backend, server, args)

	events := make(chan intf.InterfaceEvent, 16)
	errors := make(chan error, 16)

	if err := intf.WatchWireguardUserspaceInterfaces(events, errors); err != nil {
		logger.Error("Failed to watch userspace interfaces", zap.Error(err))
		return
	}

	if err := intf.WatchWireguardKernelInterfaces(events, errors); err != nil {
		logger.Error("Failed to watch kernel interfaces", zap.Error(err))
		return
	}

	logger.Debug("Starting initial interface sync")
	interfaces.SyncAll(client, backend, server, args)

	ticker := time.NewTicker(args.WatchInterval)

out:
	for {
		select {
		// We still a need periodic sync we can not (yet) monitor Wireguard interfaces
		// for changes via a netlink socket (patch is pending)
		case <-ticker.C:
			logger.Debug("Starting periodic interface sync")
			interfaces.SyncAll(client, backend, server, args)

			backend.Tick()

		case event := <-events:
			logger.Debug("Received interface event", zap.String("event", event.String()))
			interfaces.SyncAll(client, backend, server, args)

		case err := <-errors:
			logger.Error("Failed to watch for interface changes", zap.Error(err))

		case sig := <-signals:
			logger.Debug("Received signal", zap.String("signal", sig.String()))
			switch sig {
			case syscall.SIGUSR1:
				interfaces.SyncAll(client, backend, server, args)
			default:
				break out
			}
		}
	}
}
