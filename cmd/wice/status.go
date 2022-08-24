package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"riasc.eu/wice/pkg/pb"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current status of ɯice daemon",
	Run:   status,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(rootCmd, statusCmd)
}

func status(cmd *cobra.Command, args []string) {
	sts, err := client.GetStatus(context.Background(), &pb.Empty{})
	if err != nil {
		logger.Fatal("Failed to retrieve status from daemon", zap.Error(err))
	}

	mo := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		AllowPartial:    true,
		UseProtoNames:   true,
		EmitUnpopulated: false,
	}

	buf, err := mo.Marshal(sts)
	if err != nil {
		logger.Fatal("Failed to marshal", zap.Error(err))
	}

	if _, err = os.Stdout.Write(buf); err != nil {
		logger.Fatal("Failed to write to stdout", zap.Error(err))
	}
}
