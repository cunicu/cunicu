// Package main implements the command line interface
package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/config"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/util"
	"github.com/stv0g/cunicu/pkg/util/terminal"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

Author:
  Steffen Vogel <post@steffenvogel.de>

Website:
  https://cunicu.li

Code & Issues:
  https://github.com/stv0g/cunicu
`
)

var (
	logger *zap.Logger

	rootCmd = &cobra.Command{
		Use:   "cunicu",
		Short: "cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.",
		Long: Banner(terminal.IsATTY(os.Stdout)) + `cunīcu is a user-space daemon managing WireGuard® interfaces to
establish peer-to-peer connections in harsh network environments.

It relies on the awesome pion/ice package for the interactive
connectivity establishment as well as bundles the Go user-space
implementation of WireGuard in a single binary for environments
in which WireGuard kernel support has not landed yet.`,

		DisableAutoGenTag: true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
	}

	logLevel       = config.Level{Level: zapcore.InfoLevel}
	verbosityLevel int
	logFile        string
	colorMode      string
	color          bool
	stdout         io.Writer
)

func init() {
	rootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(onInitialize)

	f := rootCmd.Flags()
	f.SortFlags = false

	pf := rootCmd.PersistentFlags()
	pf.IntVarP(&verbosityLevel, "verbose", "v", 0, "verbosity level")
	pf.VarP(&logLevel, "log-level", "d", "log level (one of: debug, info, warn, error, dpanic, panic, and fatal)")
	pf.StringVarP(&logFile, "log-file", "l", "", "path of a file to write logs to")
	pf.StringVarP(&colorMode, "color", "q", "auto", "Enable colorization of output (one of: auto, always, never)")

	daemonCmd.RegisterFlagCompletionFunc("log-level", cobra.FixedCompletions([]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}, cobra.ShellCompDirectiveNoFileComp))
	daemonCmd.RegisterFlagCompletionFunc("color", cobra.FixedCompletions([]string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp))

	flagName := "output"
	daemonCmd.MarkFlagFilename(flagName, "yaml", "json")
}

func onInitialize() {
	// Initialize PRNG
	util.SetupRand()

	// Handle color output
	switch colorMode {
	case "auto":
		color = terminal.IsATTY(os.Stdout)
	case "always":
		color = true
	case "never":
		color = false
	}

	stdout = os.Stdout
	if !color {
		stdout = terminal.NewANSIStripper(stdout)
	}

	// Setup logging
	outputPaths := []string{"stdout"}
	errOutputPaths := []string{"stderr"}

	if logFile != "" {
		outputPaths = append(outputPaths, logFile)
		errOutputPaths = append(errOutputPaths, logFile)
	}

	logger = log.SetupLogging(logLevel.Level, verbosityLevel, outputPaths, errOutputPaths, color)
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
