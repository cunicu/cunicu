package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/daemon"
	"github.com/stv0g/cunicu/pkg/rpc"
	"github.com/stv0g/cunicu/pkg/util/terminal"
)

var (
	daemonCmd = &cobra.Command{
		Use:               "daemon [interface-names...]",
		Short:             "Start the daemon",
		Example:           `$ cunicu daemon -U -x mysecretpass wg0`,
		Run:               daemonRun,
		ValidArgsFunction: interfaceValidArgs,
	}

	cfg *config.Config
)

func init() {
	f := daemonCmd.Flags()
	f.SortFlags = false

	pf := daemonCmd.PersistentFlags()

	cfg = config.New(pf)

	if err := daemonCmd.RegisterFlagCompletionFunc("ice-candidate-type", cobra.FixedCompletions([]string{"host", "srflx", "prflx", "relay"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := daemonCmd.RegisterFlagCompletionFunc("ice-network-type", cobra.FixedCompletions([]string{"udp4", "udp6", "tcp4", "tcp6"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := daemonCmd.MarkPersistentFlagFilename("config", "yaml", "json"); err != nil {
		panic(err)
	}

	pf.VisitAll(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			if err := daemonCmd.RegisterFlagCompletionFunc(f.Name, BooleanCompletions); err != nil {
				panic(err)
			}
		}
	})

	rootCmd.AddCommand(daemonCmd)
}

func daemonRun(cmd *cobra.Command, args []string) {
	if _, err := io.WriteString(os.Stdout, Banner(color)); err != nil {
		logger.Fatal("Failed to write banner", zap.Error(err))
	}

	if err := cfg.Init(args); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}

	if logger.Core().Enabled(zap.DebugLevel) {
		logger.Debug("Loaded configuration:")
		wr := terminal.NewIndenter(&zapio.Writer{
			Log:   logger,
			Level: zap.DebugLevel,
		}, "   ")

		if err := cfg.Marshal(wr); err != nil {
			logger.Fatal("Failed to marshal config", zap.Error(err))
		}
	}

	// Create daemon
	d, err := daemon.New(cfg)
	if err != nil {
		logger.Fatal("Failed to create daemon", zap.Error(err))
	}

	// Create control socket server to manage daemon
	s, err := rpc.NewServer(d, cfg.RPC.Socket)
	if err != nil {
		logger.Fatal("Failed to initialize control socket", zap.Error(err))
	}

	// Delay startup until control socket client has un-waited the daemon
	if cfg.RPC.Wait {
		s.Wait()
	}

	// Blocks until stopped
	if err := d.Run(); err != nil {
		logger.Fatal("Failed start daemon", zap.Error(err))
	}

	if err := s.Close(); err != nil {
		logger.Fatal("Failed to close server", zap.Error(err))
	}

	if err := d.Close(); err != nil {
		logger.Fatal("Failed to stop daemon", zap.Error(err))
	}
}
