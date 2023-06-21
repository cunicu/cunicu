// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package main implements the command line interface
package main

import (
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/tty"
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
	logFilter string
	logFile   string
	colorMode string
}

var (
	logger *log.Logger //nolint:gochecknoglobals
	color  bool        //nolint:gochecknoglobals
	stdout io.Writer   //nolint:gochecknoglobals

	rootCmd = &cobra.Command{ //nolint:gochecknoglobals
		Use:   "cunicu",
		Short: "cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.",
		Long: Banner(tty.IsATTY(os.Stdout)) + `cunīcu is a user-space daemon managing WireGuard® interfaces to
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
	opts := &options{}

	rootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(func() {
		onInitialize(opts)
	})

	f := rootCmd.Flags()
	f.SortFlags = false

	pf := rootCmd.PersistentFlags()
	pf.StringVarP(&opts.logFilter, "log-level", "d", "info", "log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal)")
	pf.StringVarP(&opts.logFile, "log-file", "l", "", "path of a file to write logs to")
	pf.StringVarP(&opts.colorMode, "color", "q", "auto", "Enable colorization of output (one of: auto, always, never)")

	if err := rootCmd.RegisterFlagCompletionFunc("log-level", cobra.FixedCompletions([]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := rootCmd.RegisterFlagCompletionFunc("color", cobra.FixedCompletions([]string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}
}

func onInitialize(opts *options) {
	// Handle color output
	switch opts.colorMode {
	case "auto":
		color = tty.IsATTY(os.Stdout)
	case "always":
		color = true
	case "never":
		color = false
	}

	stdout = os.Stdout
	if !color {
		stdout = tty.NewANSIStripper(stdout)
	}

	// Setup logging
	outputPaths := []string{"stdout"}

	if opts.logFile != "" {
		outputPaths = append(outputPaths, opts.logFile)
	}

	var err error
	logger, err = log.SetupLogging(opts.logFilter, outputPaths, color)
	if err != nil {
		panic(err)
	}
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
