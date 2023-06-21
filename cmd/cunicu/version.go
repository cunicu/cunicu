// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/proto"
	"github.com/stv0g/cunicu/pkg/rpc"
)

type versionOptions struct {
	short  bool
	format config.OutputFormat
}

func init() { //nolint:gochecknoinits
	opts := &versionOptions{
		format: config.OutputFormatHuman,
	}

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version of the cunīcu binary and optionally also a running daemon",
		Example: `$ sudo cunicu version
client: v0.1.2 (os=linux, arch=arm64, commit=b22ee3e7, branch=master, built-at=2022-09-09T13:44:22+02:00, built-by=goreleaser)
daemon: v0.1.2 (os=linux, arch=arm64, commit=b22ee3e7, branch=master, built-at=2022-09-09T13:44:22+02:00, built-by=goreleaser)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return version(cmd, args, opts)
		},
		Args: cobra.NoArgs,
	}

	pf := cmd.PersistentFlags()
	pf.VarP(&opts.format, "format", "f", "Output `format` (one of: human, json)")
	pf.BoolVarP(&opts.short, "short", "s", false, "Only show version and nothing else")

	rootCmd.AddCommand(cmd)
}

func version(_ *cobra.Command, _ []string, opts *versionOptions) error {
	var err error

	buildInfos := &proto.BuildInfos{
		Client: buildinfo.BuildInfo(),
	}

	if rpc.DaemonRunning(rpcSockPath) {
		if rpcClient, err = rpc.Connect(rpcSockPath); err != nil {
			return fmt.Errorf("failed to connect to control socket: %w", err)
		}

		if buildInfos.Daemon, err = rpcClient.DaemonClient.GetBuildInfo(context.Background(), &proto.Empty{}); err != nil {
			logger.Fatal("Failed to retrieve status from daemon", zap.Error(err))
		}
	}

	switch opts.format {
	case config.OutputFormatJSON:
		mo := protojson.MarshalOptions{
			AllowPartial:    true,
			UseProtoNames:   true,
			EmitUnpopulated: false,
			Multiline:       true,
			Indent:          "  ",
		}

		buf, err := mo.Marshal(buildInfos)
		if err != nil {
			logger.Fatal("Failed to marshal", zap.Error(err))
		}

		fmt.Print(string(buf))
		fmt.Println()
	case config.OutputFormatHuman:
		if opts.short {
			fmt.Println(buildInfos.Client.Version)
		} else {
			fmt.Print(buildInfos.ToString())
		}

	case config.OutputFormatLogger:
	}

	return nil
}
