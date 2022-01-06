package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/pkg/socket"
)

var client *socket.Client = nil
var logger *zap.Logger

func pre(cmd *cobra.Command, args []string) error {
	var err error

	internal.SetupRand()

	logger = internal.SetupLogging()

	if client, err = socket.Connect(sockPath); err != nil {
		return fmt.Errorf("failed to connect to control socket: %w", err)
	}

	return nil
}

func post(cmd *cobra.Command, args []string) error {
	if err := client.Close(); err != nil {
		return err
	}

	if err := logger.Sync(); err != nil {
		return err
	}

	return nil
}

func main() {
	rootCmd.Execute()
}
