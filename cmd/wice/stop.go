package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/proto"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Shutdown the ɯice daemon",
	RunE:  stop,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(rootCmd, stopCmd)
}

func stop(cmd *cobra.Command, args []string) error {
	if _, err := rpcClient.Stop(context.Background(), &proto.Empty{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}
