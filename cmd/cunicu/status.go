// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/crypto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

type statusOptions struct {
	indent bool
	format config.OutputFormat
}

func init() { //nolint:gochecknoinits
	opts := &statusOptions{
		format: config.OutputFormatHuman,
	}

	cmd := &cobra.Command{
		Use:     "status [interface-name [peer-public-key]]",
		Short:   "Show current status of the cunīcu daemon, its interfaces and peers",
		Aliases: []string{"show"},
		Run: func(cmd *cobra.Command, args []string) {
			status(cmd, args, opts)
		},
		Args:              cobra.RangeArgs(0, 2),
		ValidArgsFunction: interfaceValidArgs,
	}

	pf := cmd.PersistentFlags()
	pf.VarP(&opts.format, "format", "f", "Output `format` (one of: human, json)")
	pf.BoolVarP(&opts.indent, "indent", "i", true, "Format and indent JSON output")

	if err := cmd.RegisterFlagCompletionFunc("format", cobra.FixedCompletions([]string{"human", "json"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	addClientCommand(rootCmd, cmd)
}

func status(_ *cobra.Command, args []string, opts *statusOptions) {
	p := &rpcproto.GetStatusParams{}

	if len(args) > 0 {
		p.Interface = args[0]
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

	switch opts.format {
	case config.OutputFormatJSON:
		mo := protojson.MarshalOptions{
			AllowPartial:    true,
			UseProtoNames:   true,
			EmitUnpopulated: false,
		}

		if opts.indent {
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
		if err := sts.Dump(stdout, logger.Level()); err != nil {
			logger.Fatal("Failed to write to stdout", zap.Error(err))
		}

	case config.OutputFormatLogger:
	}
}
