package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/pb"
)

var shutdownCmd = &cobra.Command{
	Use:   "shutdown",
	Short: "Shutdown the WICE daemon",
	RunE:  shutdown,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(shutdownCmd)
}

func shutdown(cmd *cobra.Command, args []string) error {
	rerr, err := client.Shutdown(context.Background(), &pb.ShutdownParams{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() {
		return fmt.Errorf("received RPC error: %w", rerr)
	}

	return nil
}
