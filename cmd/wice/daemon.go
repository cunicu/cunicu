package main

import (
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/pkg/intf"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/socket"
)

var (
	daemonCmd = &cobra.Command{
		Use:               "daemon [interfaces...]",
		Short:             "Start the daemon",
		Run:               daemon,
		PreRun:            daemonPre,
		ValidArgsFunction: daemonArgs,
	}

	cfg *config.Config
)

func init() {
	pf := daemonCmd.PersistentFlags()
	cfg = config.NewConfig(pf)

	rootCmd.AddCommand(daemonCmd)
}

func daemonPre(cmd *cobra.Command, args []string) {
	cfg.Setup()
}

func daemonArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Create Wireguard netlink socket
	client, err := wgctrl.New()
	if err != nil {
		logger.Fatal("Failed to create Wireguard client", zap.Error(err))
	}

	devs, err := client.Devices()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError | cobra.ShellCompDirectiveNoFileComp
	}

	var existing = map[string]interface{}{}
	var ifnames = []string{}

	for _, arg := range args {
		existing[arg] = nil
	}

	for _, dev := range devs {
		if _, exists := existing[dev.Name]; !exists {
			ifnames = append(ifnames, dev.Name)
		}
	}

	return ifnames, cobra.ShellCompDirectiveNoFileComp
}

func daemon(cmd *cobra.Command, args []string) {
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

	ifEvents := make(chan intf.InterfaceEvent, 16)
	errors := make(chan error, 16)

	if err := intf.WatchWireguardUserspaceInterfaces(ifEvents, errors); err != nil {
		logger.Error("Failed to watch userspace interfaces", zap.Error(err))
		return
	}

	if err := intf.WatchWireguardKernelInterfaces(ifEvents, errors); err != nil {
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

		case req := <-server.Requests:
			case *pb.StopParams:
				logger.Debug("Shutdown requested via control socket")
				break out

			case *pb.SyncInterfaceParams:
				logger.Debug("Starting interface sync triggerd via control socket")
				interfaces.SyncAll(client, backend, server, cfg)

			default:
				logger.Warn("Unhandled request", zap.Any("request", req))
			}

		case event := <-ifEvents:
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
