package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
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
	rerr, err := client.Stop(context.Background(), &pb.StopParams{})

	// We ignore ECONNRESET here since a stopped server resets the connection
	if err != nil && !errors.Is(err, unix.ECONNRESET) {
		return fmt.Errorf("failed RPC request: %w", err)
	} else if !rerr.Ok() {
		return fmt.Errorf("received RPC error: %w", rerr)
	}

	return nil
}
