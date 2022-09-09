package main

import (
	"github.com/spf13/cobra"
)

var (
	wgCmd = &cobra.Command{
		Use:   "wg",
		Short: "WireGuard commands",
		Long: `The wg sub-command mimics the wg(8) commands of the wireguard-tools package.
In contrast to the wg(8) command, the cunico sub-command delegates it tasks to a running cunucu daemon.

Currently, only a subset of the wg(8) are supported.`,
		Args: cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(wgCmd)
}
