package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/proto"
	"github.com/stv0g/cunicu/pkg/rpc"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show current status of the cunīcu daemon, its interfaces and peers",
		RunE:  version,
		Args:  cobra.NoArgs,
	}

	short bool
)

func init() {
	pf := versionCmd.PersistentFlags()
	pf.VarP(&format, "format", "f", "Output `format` (one of: human, json)")
	pf.BoolVarP(&short, "short", "s", false, "Only show version and nothing else")

	rootCmd.AddCommand(versionCmd)
}

func version(cmd *cobra.Command, args []string) error {
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

	switch format {
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
		if short {
			fmt.Println(buildInfos.Client.Version)
		} else {
			fmt.Print(buildInfos.ToString())
		}
	}

	return nil
}
