// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	statusx "google.golang.org/grpc/status"

	"github.com/stv0g/cunicu/pkg/proto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

func init() { //nolint:gochecknoinits
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration of a running cunīcu daemon.",
		Long: `
`,
	}

	setCmd := &cobra.Command{
		Use:               "set key value",
		Short:             "Update the value of a configuration setting",
		Run:               set,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: rpcValidArgs,
	}

	getCmd := &cobra.Command{
		Use:               "get [key]",
		Short:             "Get current value of a configuration setting",
		Run:               get,
		Args:              cobra.RangeArgs(0, 1),
		ValidArgsFunction: rpcValidArgs,
	}

	reloadCmd := &cobra.Command{
		Use:   "reload",
		Short: "Reload the configuration of the cunīcu daemon",
		RunE:  reload,
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(setCmd)
	cmd.AddCommand(getCmd)
	cmd.AddCommand(reloadCmd)

	addClientCommand(rootCmd, cmd)
}

func set(_ *cobra.Command, args []string) {
	key := args[0]
	values := args[1:]

	var settingValue rpcproto.ConfigValue
	if len(values) == 1 {
		settingValue.Scalar = values[0]
	} else if len(values) > 1 {
		settingValue.List = values
	}

	if _, err := rpcClient.SetConfig(context.Background(), &rpcproto.SetConfigParams{
		Settings: map[string]*rpcproto.ConfigValue{
			key: &settingValue,
		},
	}); err != nil {
		handleError(zap.FatalLevel, "Failed to set configuration", err)
	}
}

func get(_ *cobra.Command, args []string) {
	params := &rpcproto.GetConfigParams{}

	if len(args) > 0 {
		params.KeyFilter = args[0]
	}

	resp, err := rpcClient.GetConfig(context.Background(), params)
	if err != nil {
		logger.Fatal("Failed to get configuration", zap.Error(err))
	}

	keys := maps.Keys(resp.Settings)
	slices.Sort(keys)

	for _, key := range keys {
		val := resp.Settings[key]
		if val == nil {
			continue
		}

		if val.Scalar != "" {
			fmt.Printf("%s\t%s\n", key, val.Scalar)
		} else if len(val.List) > 0 {
			fmt.Printf("%s\t%s\n", key, strings.Join(val.List, "\t"))
		}
	}
}

func reload(_ *cobra.Command, _ []string) error {
	if _, err := rpcClient.ReloadConfig(context.Background(), &proto.Empty{}); err != nil {
		handleError(zap.FatalLevel, "Failed to reload configuration", err)
	}

	return nil
}

func handleError(lvl zapcore.Level, msg string, err error) {
	var field zap.Field

	if sts, ok := statusx.FromError(err); ok {
		errs := []string{}
		for _, detail := range sts.Details() {
			if err, ok := detail.(*proto.Error); ok {
				errs = append(errs, err.Message)
			}
		}

		if len(errs) > 0 {
			field = zap.Strings("errors", errs)
		} else {
			field = zap.String("error", sts.Message())
		}
	} else {
		field = zap.Error(err)
	}

	logger.Log(lvl, msg, field)
}
