// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"cunicu.li/cunicu/pkg/config"
	"cunicu.li/cunicu/pkg/rpc"
)

var (
	rpcClient   *rpc.Client //nolint:gochecknoglobals
	rpcSockPath string      //nolint:gochecknoglobals
)

func addClientCommand(rcmd, cmd *cobra.Command) {
	cmd.PersistentPreRunE = rpcConnect
	cmd.PersistentPostRunE = rpcDisconnect

	pf := cmd.PersistentFlags()
	pf.StringVarP(&rpcSockPath, "rpc-socket", "s", config.DefaultSocketPath, "Unix control and monitoring socket")

	rcmd.AddCommand(cmd)
}

func rpcConnect(_ *cobra.Command, _ []string) error {
	var err error

	if rpcClient, err = rpc.Connect(rpcSockPath); err != nil {
		return fmt.Errorf("failed to connect to control socket: %w", err)
	}

	return nil
}

func rpcDisconnect(_ *cobra.Command, _ []string) error {
	return rpcClient.Close()
}
