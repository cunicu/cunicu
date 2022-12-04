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

type options struct {
	logLevel       config.Level
	verbosityLevel int
	logFile        string
	colorMode      string
}

var (
	logger *zap.Logger //nolint:gochecknoglobals
	color  bool        //nolint:gochecknoglobals
	stdout io.Writer   //nolint:gochecknoglobals

	rootCmd = &cobra.Command{ //nolint:gochecknoglobals
		Use:   "cunicu",
		Short: "cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.",
		Long: Banner(terminal.IsATTY(os.Stdout)) + `cunīcu is a user-space daemon managing WireGuard® interfaces to
establish peer-to-peer connections in harsh network environments.

It relies on the awesome pion/ice package for the interactive
connectivity establishment as well as bundles the Go user-space
implementation of WireGuard in a single binary for environments
in which WireGuard kernel support has not landed yet.`,

		DisableAutoGenTag: true,
		SilenceUsage:      true,
		Args:              cobra.NoArgs,
		ValidArgsFunction: cobra.NoFileCompletions,
	}
)

func init() { //nolint:gochecknoinits
	opts := &options{
		logLevel: config.Level{
			Level: zapcore.InfoLevel,
		},
	}

	rootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(func() {
		onInitialize(opts)
	})

	f := rootCmd.Flags()
	f.SortFlags = false

	pf := rootCmd.PersistentFlags()
	pf.IntVarP(&opts.verbosityLevel, "verbose", "v", 0, "verbosity level")
	pf.VarP(&opts.logLevel, "log-level", "d", "log level (one of: debug, info, warn, error, dpanic, panic, and fatal)")
	pf.StringVarP(&opts.logFile, "log-file", "l", "", "path of a file to write logs to")
	pf.StringVarP(&opts.colorMode, "color", "q", "auto", "Enable colorization of output (one of: auto, always, never)")

	if err := rootCmd.RegisterFlagCompletionFunc("log-level", cobra.FixedCompletions([]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := rootCmd.RegisterFlagCompletionFunc("color", cobra.FixedCompletions([]string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := rootCmd.MarkFlagFilename("output", "yaml", "json"); err != nil {
		panic(err)
	}
}

func onInitialize(opts *options) {
	// Initialize PRNG
	util.SetupRand()

	// Handle color output
	switch opts.colorMode {
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

	if opts.logFile != "" {
		outputPaths = append(outputPaths, opts.logFile)
		errOutputPaths = append(errOutputPaths, opts.logFile)
	}

	logger = log.SetupLogging(opts.logLevel.Level, opts.verbosityLevel, outputPaths, errOutputPaths, color)
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
