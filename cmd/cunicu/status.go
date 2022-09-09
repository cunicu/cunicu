package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var (
	indent bool

	statusCmd = &cobra.Command{
		Use:               "status [interface-name [peer-public-key]]",
		Short:             "Show current status of the cunīcu daemon, its interfaces and peers",
		Aliases:           []string{"show"},
		Run:               status,
		Args:              cobra.RangeArgs(0, 2),
		ValidArgsFunction: statusValidArgs,
	}
)

func init() {
	pf := statusCmd.PersistentFlags()
	pf.VarP(&format, "format", "f", "Output `format` (one of: human, json)")
	pf.BoolVarP(&indent, "indent", "i", true, "Format and indent JSON ouput")

	daemonCmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions([]string{"human", "json"}, cobra.ShellCompDirectiveNoFileComp))

	addClientCommand(rootCmd, statusCmd)
}

func statusValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Establish RPC connection
	rpcConnect(cmd, args)
	defer rpcDisconnect(cmd, args)

	p := &rpcproto.StatusParams{}

	if len(args) > 0 {
		p.Intf = args[0]
	}

	sts, err := rpcClient.GetStatus(context.Background(), p)
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	comps := []string{}

	for _, i := range sts.Interfaces {
		if len(args) == 0 {
			comps = append(comps, i.Name)
		} else {
			for _, p := range i.Peers {
				pk, _ := crypto.ParseKeyBytes(p.PublicKey)
				comps = append(comps, pk.String())
			}
		}
	}

	return comps, cobra.ShellCompDirectiveNoFileComp
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
