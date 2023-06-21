// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	rpcproto "github.com/stv0g/cunicu/pkg/proto/rpc"
)

var errNoSuchInterface = errors.New("unknown interface")

func init() { //nolint:gochecknoinits
	cmd := &cobra.Command{
		Use:   "showconf interface-name",
		Short: "Shows the current configuration and information of the provided WireGuard interface",
		Long:  "Shows the current configuration of `interface-name` in the wg(8) format.",

		RunE:              wgShowConf,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: interfaceValidArgs,
	}

	addClientCommand(wgCmd, cmd)
}

func wgShowConf(_ *cobra.Command, args []string) error {
	intfName := args[0]

	sts, err := rpcClient.GetStatus(context.Background(), &rpcproto.GetStatusParams{
		Interface: intfName,
	})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	if len(sts.Interfaces) != 1 {
		return fmt.Errorf("%w: %s", errNoSuchInterface, intfName)
	}

	intf := sts.Interfaces[0]

	return intf.Device().Config().Dump(stdout)
}
