// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stv0g/cunicu/pkg/proto"
)

func init() { //nolint:gochecknoinits
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the cunīcu daemon",
		RunE:  restart,
		Args:  cobra.NoArgs,
	}

	addClientCommand(rootCmd, cmd)
}

func restart(_ *cobra.Command, _ []string) error {
	if _, err := rpcClient.DaemonClient.Restart(context.Background(), &proto.Empty{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}
