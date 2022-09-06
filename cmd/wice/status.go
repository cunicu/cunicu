package main

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/crypto"

	rpcproto "riasc.eu/wice/pkg/proto/rpc"
)

var (
	indent bool

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
	pf.BoolVarP(&indent, "indent", "i", true, "Format and indent JSON ouput")

	addClientCommand(rootCmd, statusCmd)
}

func status(cmd *cobra.Command, args []string) {
	p := &rpcproto.StatusParams{}

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

		if _, err = stdout.Write(buf); err != nil {
			logger.Fatal("Failed to write to stdout", zap.Error(err))
		}

	case config.OutputFormatHuman:
		sts.Dump(stdout, verbosityLevel)
	}
}
