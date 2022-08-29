package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/util"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor the ɯice daemon for events",
	Run:   monitor,
	Args:  cobra.NoArgs,
}

var format config.OutputFormat = config.OutputFormatHuman

func init() {
	addClientCommand(rootCmd, monitorCmd)

	f := monitorCmd.PersistentFlags()
	f.VarP(&format, "format", "f", fmt.Sprintf("Output `format` (one of: json, logger, human)"))
}

func monitor(cmd *cobra.Command, args []string) {
	signals := util.SetupSignals()

	logger := logger.Named("events")

	mo := protojson.MarshalOptions{
		UseProtoNames: true,
	}

out:
	for {
		select {
		case sig := <-signals:
			logger.Debug("Received signal", zap.Any("signal", sig))
			break out

		case evt := <-rpcClient.Events:
			switch format {
			case config.OutputFormatJSON:
				buf, err := mo.Marshal(evt)
				if err != nil {
					logger.Fatal("Failed to marshal", zap.Error(err))
				}
				buf = append(buf, '\n')

				if _, err = stdout.Write(buf); err != nil {
					logger.Fatal("Failed to write to stdout", zap.Error(err))
				}

			case config.OutputFormatHuman:
				fallthrough
			case config.OutputFormatLogger:
				evt.Log(logger, "Event")
			}
		}
	}
}
