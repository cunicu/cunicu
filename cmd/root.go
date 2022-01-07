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

	// set via ldflags -X / goreleaser
	version string
	commit  string
	date    string
)

func init() {
	cobra.OnInitialize(
		internal.SetupRand,
		setupLogging,
	)

	rootCmd.Version = version
}

func setupLogging() {
	logger = internal.SetupLogging()
}
