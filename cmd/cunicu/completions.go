// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"cunicu.li/cunicu/pkg/crypto"
	rpcproto "cunicu.li/cunicu/pkg/proto/rpc"
	"cunicu.li/cunicu/pkg/rpc"
)

//nolint:gochecknoglobals
var BooleanCompletions = cobra.FixedCompletions([]string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp)

func interfaceValidArgs(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	rpcClient, err := rpc.Connect(rpcSockPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer rpcClient.Close()

	p := &rpcproto.GetStatusParams{}
	if len(args) > 0 {
		p.Interface = args[0]
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

func getCommandParts(cmd *cobra.Command) []string {
	var chain []string
	if parent := cmd.Parent(); parent != nil {
		chain = getCommandParts(parent)
	}

	parts := strings.SplitN(cmd.Use, " ", 2)
	chain = append(chain, parts[0])

	return chain
}

func rpcValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	rpcClient, err := rpc.Connect(rpcSockPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	defer rpcClient.Close()

	resp, err := rpcClient.GetCompletion(context.Background(), &rpcproto.GetCompletionParams{
		Cmd:        getCommandParts(cmd),
		Args:       args,
		ToComplete: toComplete,
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return resp.Options, cobra.ShellCompDirective(resp.Flags)
}
