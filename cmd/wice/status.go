package main

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/util"
)

var (
	color     bool
	indent    bool
	verbosity int

	statusCmd = &cobra.Command{
		Use:   "status [flags] [intf [peer]]",
		Short: "Show current status of the ɯice daemon, its interfaces and peers",
		Run:   status,
		Args:  cobra.RangeArgs(0, 2),
	}
)

func init() {
	pf := statusCmd.PersistentFlags()
	pf.VarP(&format, "format", "f", "Output `format` (one of: human, json)")
	pf.IntVarP(&verbosity, "verbose", "v", 5, "Verbosity level for output (1-6)")
	pf.BoolVarP(&color, "color", "c", true, "Enable colorization of output")
	pf.BoolVarP(&indent, "indent", "i", true, "Format and indent JSON ouput")

	addClientCommand(rootCmd, statusCmd)
}

func status(cmd *cobra.Command, args []string) {
	p := &pb.StatusParams{}

	if len(args) > 0 {
		p.Intf = args[0]
		if len(args) > 1 {
			pk, err := crypto.ParseKey(args[1])
			if err != nil {
				logger.Fatal("Invalid public key", zap.Error(err))
			}

			p.Peer = pk.Bytes()
		}
	}

	sts, err := rpcClient.GetStatus(context.Background(), p)
	if err != nil {
		logger.Fatal("Failed to retrieve status from daemon", zap.Error(err))
	}

	var wr io.Writer = os.Stdout
	if supportsColor := util.IsATTY(); !supportsColor || !color {
		wr = &util.ANSIStripper{
			Writer: wr,
		}
	}

	switch format {
	case config.OutputFormatJSON:
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

	case config.OutputFormatHuman:
		sts.Dump(wr, verbosity)
	}
}
