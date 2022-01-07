package main

import (
	"github.com/spf13/cobra"
	"riasc.eu/wice/internal"
)

var (
	// Used for flags.
	CfgFile string

	rootCmd = &cobra.Command{
		Use:   "wice",
		Short: "Wireguard Interactive Connectitivty Establishment",
	}
)

func init() {
	cobra.OnInitialize(
		internal.SetupRand,
		setupLogging,
	)
}

func setupLogging() {
	logger = internal.SetupLogging()
}
