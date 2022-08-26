package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"riasc.eu/wice/pkg/pb"
)

var (
	wgShowConfCmd = &cobra.Command{
		Use:   "showconf [flags] <interface>",
		Short: "Shows the current configuration and device information",
		Long:  "Sets the current configuration of <interface> to the contents of <configuration-filename>, which must be in the wg(8) format.",

		RunE: wgShowConf,
		Args: cobra.ExactArgs(1),
	}
)

func init() {
	addClientCommand(wgCmd, wgShowConfCmd)
}

func wgShowConf(cmd *cobra.Command, args []string) error {
	intfName := args[0]

	sts, err := rpcClient.GetStatus(context.Background(), &pb.StatusParams{
		Intf: intfName,
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
