package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/proto"
)

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the cunicu daemon",
	RunE:  restart,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(rootCmd, restartCmd)
}

func restart(cmd *cobra.Command, args []string) error {
	if _, err := rpcClient.Restart(context.Background(), &proto.Empty{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}
