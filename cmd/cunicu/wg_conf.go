package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var (
	wgShowConfCmd = &cobra.Command{
		Use:   "showconf interface-name",
		Short: "Shows the current configuration and information of the provided WireGuard interface",
		Long:  "Shows the current configuration of `interface-name` in the wg(8) format.",

		RunE:              wgShowConf,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: interfaceValidArg,
	}
)

func init() {
	addClientCommand(wgCmd, wgShowConfCmd)
}

func wgShowConf(cmd *cobra.Command, args []string) error {
	intfName := args[0]

	sts, err := rpcClient.GetStatus(context.Background(), &rpcproto.GetStatusParams{
		Interface: intfName,
	})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	if len(sts.Interfaces) != 1 {
		return fmt.Errorf("failed to find interface '%s'", intfName)
	}

	intf := sts.Interfaces[0]

	return intf.Device().Config().Dump(stdout)
}
