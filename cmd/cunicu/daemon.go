// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/stv0g/cunicu/pkg/tty"
)

func init() { //nolint:gochecknoinits
	cmd := &cobra.Command{
		Use:   "daemon [interface-names...]",
		Short: "Start the main daemon",
		Long: `Starts the main cunicu agent.
		
Sending a SIGUSR1 signal to the daemon will trigger an immediate synchronization of all WireGuard interfaces.`,
		Example:           "$ cunicu daemon -U -x mysecretpass wg0",
		ValidArgsFunction: interfaceValidArgs,
	}

	f := cmd.Flags()
	f.SortFlags = false

	pf := cmd.PersistentFlags()

	cfg := config.New(pf)

	if err := cmd.RegisterFlagCompletionFunc("ice-candidate-type", cobra.FixedCompletions([]string{"host", "srflx", "prflx", "relay"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := cmd.RegisterFlagCompletionFunc("ice-network-type", cobra.FixedCompletions([]string{"udp4", "udp6", "tcp4", "tcp6"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := cmd.MarkPersistentFlagFilename("config", "yaml", "json"); err != nil {
		panic(err)
	}

	pf.VisitAll(func(f *pflag.Flag) {
		if f.Value.Type() == "bool" {
			if err := cmd.RegisterFlagCompletionFunc(f.Name, BooleanCompletions); err != nil {
				panic(err)
			}
		}
	})

	cmd.Run = func(cmd *cobra.Command, args []string) {
		daemonRun(cmd, args, cfg)
	}

	rootCmd.AddCommand(cmd)
}

func daemonRun(_ *cobra.Command, args []string, cfg *config.Config) {
	if _, err := io.WriteString(os.Stdout, Banner(color)); err != nil {
		logger.Fatal("Failed to write banner", zap.Error(err))
	}

	if err := cfg.Init(args); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}

	// Require experimental env var
	if !cfg.Experimental {
		logger.Fatal(`cunicu is currently under development.

	You should only be running it if you are testing/developing it.
	Please set the env var CUNICU_EXPERIMENTAL=1 to bypass this warning.
	
	Please feel free to join the development
	 - at Github: https://github.com/stv0g/cunicu
	 - via Slack: #cunicu in the Gophers workspace`)
	}

	if logger.Core().Enabled(zap.DebugLevel) {
		logger = logger.Named("config")

		logger.DebugV(1, "Loaded configuration:")
		wr := tty.NewIndenter(&zapio.Writer{
			Log:   logger.Logger,
			Level: zap.DebugLevel - 1,
		}, "   ")

		if err := cfg.Marshal(wr); err != nil {
			logger.Fatal("Failed to marshal config", zap.Error(err))
		}
	}

	// Create daemon
	d, err := daemon.NewDaemon(cfg)
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
	if err := d.Start(); err != nil {
		logger.Fatal("Failed start daemon", zap.Error(err))
	}

	if err := d.Close(); err != nil {
		logger.Fatal("Failed to stop daemon", zap.Error(err))
	}

	if err := s.Close(); err != nil {
		logger.Fatal("Failed to close server", zap.Error(err))
	}
}
