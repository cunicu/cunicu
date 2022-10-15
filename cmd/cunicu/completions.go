package main

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/crypto"
	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var BooleanCompletions = cobra.FixedCompletions([]string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp)

func interfaceValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Establish RPC connection
	if err := rpcConnect(cmd, args); err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	defer rpcDisconnect(cmd, args)

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
