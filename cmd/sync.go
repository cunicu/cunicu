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
		Short: "Synchronize interfaces",
		Long:  "Synchronizes the internal daemon state with the state of the WireGuard interfaces",
		RunE:  sync,
		Args:  cobra.RangeArgs(0, 1),
	}
)

func init() {
	addClientCommand(RootCmd, syncCmd)
}

func sync(cmd *cobra.Command, args []string) error {
	rerr, err := client.Sync(context.Background(), &pb.SyncParams{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() {
		return fmt.Errorf("received RPC error: %w", rerr)
	}

	return nil
}
