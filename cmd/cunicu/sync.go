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
		Use:   "sync",
		Short: "Synchronize cunīcu daemon state",
		Long:  "Synchronizes the internal daemon state with kernel routes, interfaces and addresses",
		RunE:  sync,
		Args:  cobra.NoArgs,
	}

	addClientCommand(rootCmd, cmd)
}

func sync(_ *cobra.Command, _ []string) error {
	_, err := rpcClient.Sync(context.Background(), &proto.Empty{})
	if err != nil {
		return fmt.Errorf("failed RPC request: %w", err)
	}
	return nil
}
