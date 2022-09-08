package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/proto"
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
	addClientCommand(rootCmd, syncCmd)
}

func sync(cmd *cobra.Command, args []string) error {
	_, err := rpcClient.Sync(context.Background(), &proto.Empty{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}
	return nil
}