package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/internal"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor the WICE daemon for events",
	RunE:  monitor,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(monitorCmd)
}

func monitor(cmd *cobra.Command, args []string) error {
	signals := internal.SetupSignals()

	logger := zap.L().Named("events")

out:
	for {
		select {
		case sig := <-signals:
			logger.Info("Received signal", zap.Any("signal", sig))
			break out

		case evt := <-client.Events:
			evt.Log(logger, "Event")
		}
	}

	return nil
}
