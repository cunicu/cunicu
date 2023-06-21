// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/config"
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
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: validConfigSettings,
	}

	getCmd := &cobra.Command{
		Use:               "get [key]",
		Short:             "Get current value of a configuration setting",
		Run:               get,
		Args:              cobra.RangeArgs(0, 1),
		ValidArgsFunction: validConfigSettings,
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

func getCompletions(typ reflect.Type, haveCompleted, toComplete string) ([]string, cobra.ShellCompDirective) {
	tagComplete := strings.Split(toComplete, ".")[0]

	flags := cobra.ShellCompDirectiveNoFileComp
	fields := []reflect.StructField{}
	comps := []string{}
	structComps := []string{}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tagLine := field.Tag.Get("yaml")
		comp := strings.Split(tagLine, ",")[0]

		if strings.HasPrefix(comp, tagComplete) {
			if field.Type.Kind() == reflect.Struct {
				comp += "."
				flags |= cobra.ShellCompDirectiveNoSpace

				structComps = append(structComps, comp)
			}

			fields = append(fields, field)
			comps = append(comps, haveCompleted+comp)
		}
	}

	if len(fields) == 1 && fields[0].Type.Kind() == reflect.Struct {
		if strings.HasPrefix(toComplete, structComps[0]) {
			toComplete = strings.TrimPrefix(toComplete, structComps[0])
		} else {
			toComplete = ""
		}

		return getCompletions(fields[0].Type, haveCompleted+structComps[0], toComplete)
	}

	return comps, flags
}

func validConfigSettings(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	t := reflect.TypeOf(config.Settings{})

	return getCompletions(t, "", toComplete)
}

func set(_ *cobra.Command, args []string) {
	settings := map[string]string{
		args[0]: args[1],
	}

	if _, err := rpcClient.SetConfig(context.Background(), &rpcproto.SetConfigParams{
		Settings: settings,
	}); err != nil {
		logger.Fatal("Failed to set configuration", zap.Error(err))
	}
}

func get(_ *cobra.Command, args []string) {
	params := &rpcproto.GetConfigParams{}

	if len(args) > 0 {
		params.KeyFilter = args[0]
	}

	resp, err := rpcClient.GetConfig(context.Background(), params)
	if err != nil {
		logger.Fatal("Failed to set configuration", zap.Error(err))
	}

	keys := maps.Keys(resp.Settings)
	slices.Sort(keys)

	for _, key := range keys {
		fmt.Printf("%s\t%s\n", key, resp.Settings[key])
	}
}

func reload(_ *cobra.Command, _ []string) error {
	if _, err := rpcClient.ReloadConfig(context.Background(), &proto.Empty{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}
