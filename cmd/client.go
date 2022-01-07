package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/pkg/socket"
)

var (
	client   *socket.Client = nil
	sockPath string
)

func addClientCommand(cmd *cobra.Command) {
	cmd.PreRunE = connect
	cmd.PostRunE = disconnect

	pf := cmd.PersistentFlags()

	pf.StringVarP(&sockPath, "socket", "s", config.DefaultSocketPath, "Unix control and monitoring socket")

	rootCmd.AddCommand(cmd)
}

func connect(cmd *cobra.Command, args []string) error {
	var err error

	if client, err = socket.Connect(sockPath); err != nil {
		return fmt.Errorf("failed to connect to control socket: %w", err)
	}

	return nil
}

func disconnect(cmd *cobra.Command, args []string) error {
	if err := client.Close(); err != nil {
		return err
	}

	return nil
}
