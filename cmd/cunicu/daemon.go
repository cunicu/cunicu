package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	cunicu "github.com/stv0g/cunicu/pkg"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/rpc"
	"github.com/stv0g/cunicu/pkg/util/terminal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

var (
	daemonCmd = &cobra.Command{
		Use:               "daemon [interface-names...]",
		Short:             "Start the daemon",
		Example:           `$ cunicu daemon -u -x mysecretpass wg0`,
		Run:               daemon,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	cfg *config.Config
)

func init() {
	f := daemonCmd.Flags()
	f.SortFlags = false

	pf := daemonCmd.PersistentFlags()

	cfg = config.New(pf)

	daemonCmd.RegisterFlagCompletionFunc("ice-candidate-type", cobra.FixedCompletions([]string{"host", "srflx", "prflx", "relay"}, cobra.ShellCompDirectiveNoFileComp))
	daemonCmd.RegisterFlagCompletionFunc("ice-network-type", cobra.FixedCompletions([]string{"udp4", "udp6", "tcp4", "tcp6"}, cobra.ShellCompDirectiveNoFileComp))

	pf.VisitAll(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			daemonCmd.RegisterFlagCompletionFunc(f.Name, config.BooleanCompletions)
		}
	})

	rootCmd.AddCommand(daemonCmd)
}

func daemon(cmd *cobra.Command, args []string) {
	io.WriteString(os.Stdout, Banner(color))

	if err := cfg.Load(); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}

	if logger.Core().Enabled(zap.DebugLevel) {
		logger.Debug("Loaded configuration:")
		wr := terminal.NewIndenter(&zapio.Writer{
			Log:   logger,
			Level: zap.DebugLevel,
		}, "   ")
		cfg.Marshal(wr)
	}

	// Create daemon
	daemon, err := cunicu.NewDaemon(cfg)
	if err != nil {
		logger.Fatal("Failed to create daemon", zap.Error(err))
	}

	// Create control socket server to manage daemon
	svr, err := rpc.NewServer(daemon, cfg.RPC.Socket)
	if err != nil {
		logger.Fatal("Failed to initialize control socket", zap.Error(err))
	}

	// Delay startup until control socket client has un-waited the daemon
	if cfg.RPC.Wait {
		svr.Wait()
	}

	// Blocks until stopped
	if err := daemon.Run(); err != nil {
		logger.Fatal("Failed start daemon", zap.Error(err))
	}

	if err := svr.Close(); err != nil {
		logger.Fatal("Failed to close server", zap.Error(err))
	}

	if err := daemon.Close(); err != nil {
		logger.Fatal("Failed to stop daemon", zap.Error(err))
	}
}
