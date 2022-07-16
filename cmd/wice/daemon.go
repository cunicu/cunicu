package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/socket"
)

var (
	daemonCmd = &cobra.Command{
		Use:               "daemon [interfaces...]",
		Short:             "Start the daemon",
		Run:               daemon,
		ValidArgsFunction: daemonCompletionArgs,
	}

	cfg *config.Config
)

func init() {
	f := daemonCmd.Flags()
	f.SortFlags = false

	pf := daemonCmd.PersistentFlags()

	cfg = config.NewConfig(pf)

	RootCmd.AddCommand(daemonCmd)
}

func daemonCompletionArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Create Wireguard netlink socket
	client, err := wgctrl.New()
	if err != nil {
		logger.Fatal("Failed to create Wireguard client", zap.Error(err))
	}

	devs, err := client.Devices()
	if err != nil {
		return []string{}, cobra.ShellCompDirectiveError | cobra.ShellCompDirectiveNoFileComp
	}

	var existing = map[string]any{}
	var intfNames = []string{}

	for _, arg := range args {
		existing[arg] = nil
	}

	for _, dev := range devs {
		if _, exists := existing[dev.Name]; !exists {
			intfNames = append(intfNames, dev.Name)
		}
	}

	return intfNames, cobra.ShellCompDirectiveNoFileComp
}

func daemon(cmd *cobra.Command, args []string) {
	if err := cfg.Setup(args); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}

	if logger.Core().Enabled(zap.DebugLevel) {
		logger.Debug("Loaded configuration:")
		if err := cfg.Dump(&zapio.Writer{Log: logger}); err != nil {
			logger.Fatal("Failed to dump configuration", zap.Error(err))
		}
	}

	// Create daemon
	daemon, err := pkg.NewDaemon(cfg)
	if err != nil {
		logger.Fatal("Failed to create daemon", zap.Error(err))
	}

	// Create control socket server to manage daemon
	svr, err := socket.Listen("unix", cfg.Socket.Path)
	if err != nil {
		logger.Fatal("Failed to initialize control socket", zap.Error(err))
	}

	svr.RegisterDaemon(daemon)

	// Delay startup until control socket client has un-waited the daemon
	if cfg.Socket.Wait {
		svr.Wait()
	}

	if err := daemon.Run(); err != nil {
		logger.Fatal("Failed run daemon", zap.Error(err))
	}
}
