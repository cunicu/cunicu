package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/proto"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the configuration of the cunīcu daemon",
	RunE:  reload,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(rootCmd, reloadCmd)
}

func reload(cmd *cobra.Command, args []string) error {
	if _, err := rpcClient.ReloadConfig(context.Background(), &proto.Empty{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}
