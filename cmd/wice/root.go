package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap/zapcore"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/log"
	"riasc.eu/wice/pkg/util"
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
		Long:  "WireGuard Interactive Connectivity Establishment",

		DisableAutoGenTag: true,
		Version:           version,
	}

	// set via ldflags -X / goreleaser
	version string
	date    string
	// commit  string

	logLevel = config.Level{Level: zapcore.InfoLevel}
	logFile  string
)

func init() {
	rootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(
		util.SetupRand,
		setupLogging,
	)

	f := rootCmd.Flags()
	f.SortFlags = false

	pf := rootCmd.PersistentFlags()
	pf.VarP(&logLevel, "log-level", "d", "log level (one of: debug, info, warn, error, dpanic, panic, and fatal)")
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

func main() {
	if os.Args[0] == "wg" {
		if err := wgCmd.Execute(); err != nil {
			os.Exit(1)
		}
	} else {
		if err := rootCmd.Execute(); err != nil {
			os.Exit(1)
		}
	}
}
