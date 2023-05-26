// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var wgCmd = &cobra.Command{
	Use:   "wg",
	Short: "WireGuard commands",
	Long: `The wg sub-command mimics the wg(8) commands of the wireguard-tools package.
In contrast to the wg(8) command, the cunico sub-command delegates it tasks to a running cunucu daemon.

Currently, only a subset of the wg(8) are supported.`,
	Args: cobra.NoArgs,
}

func init() { //nolint:gochecknoinits
	rootCmd.AddCommand(wgCmd)
}
