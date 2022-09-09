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
		Short: "Synchronize cunīcu daemon state",
		Long:  "Synchronizes the internal daemon state with kernel routes, interfaces and addresses",
		RunE:  sync,
		Args:  cobra.NoArgs,
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
