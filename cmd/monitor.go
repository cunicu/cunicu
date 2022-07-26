package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"riasc.eu/wice/pkg"
	"riasc.eu/wice/pkg/config"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor the ɯice daemon for events",
	Run:   monitor,
	Args:  cobra.NoArgs,
}

var format config.OutputFormat

func init() {
	addClientCommand(RootCmd, monitorCmd)

	f := monitorCmd.PersistentFlags()
	f.VarP(&format, "format", "f", fmt.Sprintf("Output `format` (one of: %s)", strings.Join(config.OutputFormatNames, ", ")))
}

func monitor(cmd *cobra.Command, args []string) {
	signals := pkg.SetupSignals()

	logger := logger.Named("events")

	mo := protojson.MarshalOptions{
		UseProtoNames: true,
	}

out:
	for {
		select {
		case sig := <-signals:
			logger.Info("Received signal", zap.Any("signal", sig))
			break out

		case evt := <-client.Events:
			switch format {
			case config.OutputFormatCSV:
			case config.OutputFormatJSON:
				buf, err := mo.Marshal(evt)
				if err != nil {
					logger.Fatal("Failed to marshal", zap.Error(err))
				}
				buf = append(buf, '\n')

				if _, err = os.Stdout.Write(buf); err != nil {
					logger.Fatal("Failed to write to stdout", zap.Error(err))
				}

			case config.OutputFormatLogger:
				evt.Log(logger, "Event")
			}
		}
	}
}
