// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"cunicu.li/cunicu/pkg/config"
	"cunicu.li/cunicu/pkg/log"
	osx "cunicu.li/cunicu/pkg/os"
	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
)

type monitorOptions struct {
	format config.OutputFormat
}

func init() { //nolint:gochecknoinits
	opts := &monitorOptions{
		format: config.OutputFormatHuman,
	}

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor the cunīcu daemon for events",
		Run: func(cmd *cobra.Command, args []string) {
			monitor(cmd, args, opts)
		},
		Args: cobra.NoArgs,
	}

	addClientCommand(rootCmd, cmd)

	f := cmd.PersistentFlags()
	f.VarP(&opts.format, "format", "f", "Output `format` (one of: json, logger, human)")
}

type monitorEventHandler struct {
	opts   *monitorOptions
	mo     *protojson.MarshalOptions
	logger *log.Logger
}

func (h *monitorEventHandler) OnEvent(e *rpcproto.Event) {
	switch h.opts.format {
	case config.OutputFormatJSON:
		buf, err := h.mo.Marshal(e)
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
		e.Log(logger, "Event")
	}
}

func monitor(_ *cobra.Command, _ []string, opts *monitorOptions) {
	eh := &monitorEventHandler{
		mo: &protojson.MarshalOptions{
			UseProtoNames: true,
		},
		opts:   opts,
		logger: logger.Named("events"),
	}

	rpcClient.AddEventHandler(eh)

	for signal := range osx.SetupSignals() {
		logger.Debug("Received signal", zap.Any("signal", signal))
		break
	}
}
