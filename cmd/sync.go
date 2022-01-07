package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/pb"
)

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Synchronize interfaces with kernel or configuration files",
		Args:  cobra.NoArgs,
	}

	syncConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Synchronize the active interfaces with the on-disc configuration files",
		RunE:  syncConfig,
		Args:  cobra.NoArgs,
	}

	syncInterfacesCmd = &cobra.Command{
		Use:   "interfaces",
		Short: "Synchronize the daemons state with the existing interface",
		RunE:  syncInterfaces,
		Args:  cobra.NoArgs,
	}
)

func init() {
	syncCmd.AddCommand(syncConfigCmd)
	syncCmd.AddCommand(syncInterfacesCmd)

	addClientCommand(syncCmd)
}

func syncConfig(cmd *cobra.Command, args []string) error {
	rerr, err := client.SyncConfig(context.Background(), &pb.SyncConfigParams{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() {
		return fmt.Errorf("received RPC error: %w", rerr)
	}

	return nil
}

func syncInterfaces(cmd *cobra.Command, args []string) error {
	rerr, err := client.SyncInterfaces(context.Background(), &pb.SyncInterfaceParams{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() {
		return fmt.Errorf("received RPC error: %w", rerr)
	}

	return nil
}
