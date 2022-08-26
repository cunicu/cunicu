package main

import (
	"github.com/spf13/cobra"
)

var (
	wgCmd = &cobra.Command{
		Use:   "wg",
		Short: "WireGuard commands",
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(wgCmd)
}
