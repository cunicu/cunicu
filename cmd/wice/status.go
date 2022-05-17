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
	RunE:  status,
	Args:  cobra.NoArgs,
}

func init() {
	addClientCommand(RootCmd, statusCmd)
}

func status(cmd *cobra.Command, args []string) error {
	sts, err := client.GetStatus(context.Background(), &pb.Void{})
	if err != nil {
		logger.Fatal("Failed to retrieve status from daemon", zap.Error(err))
	}

	mo := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		AllowPartial:    true,
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}

	buf, err := mo.Marshal(sts)
	if err != nil {
		logger.Fatal("Failed to marshal", zap.Error(err))
	}

	_, err = os.Stdout.Write(buf)

	return err
}
