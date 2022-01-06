package main

import (
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/cli"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/pkg/intf"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	_ "riasc.eu/wice/pkg/signaling/k8s"
	_ "riasc.eu/wice/pkg/signaling/p2p"
	"riasc.eu/wice/pkg/socket"
)

var (
	logger *zap.Logger
	cfg    *config.Config

	rootCmd = &cobra.Command{
		Use:                "wice",
		Short:              "The WICE daemon",
		Run:                run,
		PersistentPreRunE:  pre,
		PersistentPostRunE: post,
	}
)

func main() {
	pf := rootCmd.LocalFlags()

	cfg = config.NewConfig(pf)

	cobra.OnInitialize(cfg.Setup)

	rootCmd.AddCommand(cli.NewDocsCommand(rootCmd))

	rootCmd.Execute()
}

func pre(cmd *cobra.Command, args []string) error {
	logger = internal.SetupLogging()

	return nil
}

func post(cmd *cobra.Command, args []string) error {
	if err := logger.Sync(); err != nil {
		// return err
	}

	return nil
	}

func run(cmd *cobra.Command, args []string) {
	signals := internal.SetupSignals()

	if logger.Core().Enabled(zap.DebugLevel) {
		wr := &zapio.Writer{Log: logger}
		cfg.Dump(wr)
	}

	// Create control socket server
	server, err := socket.Listen("unix", cfg.Socket, cfg.SocketWait)
	if err != nil {
		logger.Fatal("Failed to initialize control socket", zap.Error(err))
	}

	// Create backend
	var backend signaling.Backend
	if len(cfg.Backends) == 1 {
		backend, err = signaling.NewBackend(cfg.Backends[0], server)
	} else {
		backend, err = signaling.NewMultiBackend(cfg.Backends, server)
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

	if err := interfaces.CreateFromArgs(client, backend, server, cfg); err != nil {
		logger.Fatal("Failed to create interfaces", zap.Error(err))
	}

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
	interfaces.SyncAll(client, backend, server, cfg)

	ticker := time.NewTicker(cfg.WatchInterval)

out:
	for {
		select {
		// We still a need periodic sync we can not (yet) monitor Wireguard interfaces
		// for changes via a netlink socket (patch is pending)
		case <-ticker.C:
			logger.Debug("Starting periodic interface sync")
			interfaces.SyncAll(client, backend, server, cfg)

			backend.Tick()

		case event := <-events:
			logger.Debug("Received interface event", zap.String("event", event.String()))
			interfaces.SyncAll(client, backend, server, cfg)

		case err := <-errors:
			logger.Error("Failed to watch for interface changes", zap.Error(err))

		case sig := <-signals:
			logger.Debug("Received signal", zap.String("signal", sig.String()))
			switch sig {
			case syscall.SIGUSR1:
				interfaces.SyncAll(client, backend, server, cfg)
			default:
				break out
			}
		}
	}
}
