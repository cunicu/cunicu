package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/pb"
)

var (
	syncCmd = &cobra.Command{
		Use:                "sync",
		PersistentPreRunE:  pre,
		PersistentPostRunE: post,
	}

	syncConfigCmd = &cobra.Command{
		Use:  "config",
		RunE: syncConfig,
	}

	syncInterfacesCmd = &cobra.Command{
		Use:  "interfaces",
		RunE: syncInterfaces,
	}
)

func init() {
	syncCmd.AddCommand(syncConfigCmd)
	syncCmd.AddCommand(syncInterfacesCmd)
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
