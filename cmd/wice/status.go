package main

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/util"
)

var (
	json      bool
	color     bool
	indent    bool
	verbosity int
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current status of ɯice daemon",
	Run:   status,
	Args:  cobra.NoArgs,
}

func init() {
	pf := statusCmd.PersistentFlags()
	pf.BoolVarP(&json, "json", "j", false, "Format status in JSON")
	pf.IntVarP(&verbosity, "verbose", "v", 6, "Verbosity level for output (1-6)")
	pf.BoolVarP(&color, "color", "c", true, "Enable colorization of output")
	pf.BoolVarP(&indent, "indent", "i", true, "Format and indent JSON ouput")

	addClientCommand(rootCmd, statusCmd)
}

func status(cmd *cobra.Command, args []string) {
	sts, err := rpcClient.GetStatus(context.Background(), &pb.Empty{})
	if err != nil {
		logger.Fatal("Failed to retrieve status from daemon", zap.Error(err))
	}

	var wr io.Writer = os.Stdout
	if supportsColor := util.IsATTY(); !supportsColor || !color {
		wr = &util.ANSIStripper{
			Writer: wr,
		}
	}

	if json {
		mo := protojson.MarshalOptions{
			AllowPartial:    true,
			UseProtoNames:   true,
			EmitUnpopulated: false,
		}

		if indent {
			mo.Multiline = true
			mo.Indent = "  "
		}

		buf, err := mo.Marshal(sts)
		if err != nil {
			logger.Fatal("Failed to marshal", zap.Error(err))
		}

		if _, err = wr.Write(buf); err != nil {
			logger.Fatal("Failed to write to stdout", zap.Error(err))
		}
	} else {
		sts.Dump(wr, verbosity)
	}
}
