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
		Use:   "stop",
		Short: "Shutdown the cunīcu daemon",
		RunE:  stop,
		Args:  cobra.NoArgs,
	}

	addClientCommand(rootCmd, cmd)
}

func stop(_ *cobra.Command, _ []string) error {
	if _, err := rpcClient.Stop(context.Background(), &proto.Empty{}); err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}

	return nil
}
