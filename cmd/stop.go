package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/pb"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Shutdown the ɯice daemon",
	RunE:  stop,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(RootCmd, stopCmd)
}

func stop(cmd *cobra.Command, args []string) error {
	// TODO: Ignore errors caused by closed connection or gracefully shutdown the server
	if rerr, err := client.Stop(context.Background(), &pb.StopParams{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() {
		return fmt.Errorf("received RPC error: %w", rerr)
	}

	return nil
}
