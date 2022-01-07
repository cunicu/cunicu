package main

import (
	"github.com/spf13/cobra"
	"riasc.eu/wice/internal"
)

const (
	usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

Credits:
  Steffen Vogel <post@steffenvogel.de>

Website:
  https://github.com/stv0g/wice
`
)

var (
	// Used for flags.
	CfgFile string

	rootCmd = &cobra.Command{
		Use:   "wice",
		Short: "WICE",
		Long:  "Wireguard Interactive Connectitivty Establishment",

		// The main wice command is just an alias for "wice daemon"
		Run: daemon,
	}

	version string //lint:ignore U1000 set via ldflags -X / goreleaser
	commit  string //lint:ignore U1000 set via ldflags -X / goreleaser
	date    string //lint:ignore U1000 set via ldflags -X / goreleaser
)

func init() {
	cobra.OnInitialize(
		internal.SetupRand,
		setupLogging,
	)

	rootCmd.Version = version
	rootCmd.SetUsageTemplate(usageTemplate)
}

func setupLogging() {
	logger = internal.SetupLogging()
}
