package main

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"riasc.eu/wice/internal"
	"riasc.eu/wice/internal/log"
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
	rootCmd = &cobra.Command{
		Use:   "wice",
		Short: "ɯice",
		Long:  "Wireguard Interactive Connectivity Establishment",

		// The main wice command is just an alias for "wice daemon"
		Run:               daemon,
		DisableAutoGenTag: true,
		Version:           version,
	}

	version string //lint:ignore U1000 set via ldflags -X / goreleaser
	commit  string //lint:ignore U1000 set via ldflags -X / goreleaser
	date    string //lint:ignore U1000 set via ldflags -X / goreleaser

	logLevel = level{zapcore.InfoLevel}
	logFile  string
)

type level struct {
	zapcore.Level
}

func (l *level) Type() string {
	return "string"
}

func init() {
	rootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(
		internal.SetupRand,
		setupLogging,
	)

	f := rootCmd.Flags()
	f.SortFlags = false

	pf := rootCmd.PersistentFlags()
	pf.VarP(&logLevel, "log-level", "d", "log level (one of \"debug\", \"info\", \"warn\", \"error\", \"dpanic\", \"panic\", and \"fatal\")")
	pf.StringVarP(&logFile, "log-file", "l", "", "path of a file to write logs to")
}

func setupLogging() {
	outputPaths := []string{"stdout"}
	errOutputPaths := []string{"stderr"}

	if logFile != "" {
		outputPaths = append(outputPaths, logFile)
		errOutputPaths = append(errOutputPaths, logFile)
	}

	logger = log.SetupLogging(logLevel.Level, outputPaths, errOutputPaths)
}
