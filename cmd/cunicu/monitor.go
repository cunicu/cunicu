package main

import (
	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor the cunicu daemon for events",
	Run:   monitor,
	Args:  cobra.NoArgs,
}

var format config.OutputFormat = config.OutputFormatHuman

func init() {
	addClientCommand(rootCmd, monitorCmd)

	f := monitorCmd.PersistentFlags()
	f.VarP(&format, "format", "f", "Output `format` (one of: json, logger, human)")
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
