// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package main implements the command line interface
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"cunicu.li/cunicu/pkg/config"
	"cunicu.li/cunicu/pkg/log"
	"cunicu.li/cunicu/pkg/tty"
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
  https://github.com/cunicu/cunicu
`
)

type options struct {
	logRules  []string
	logFiles  []string
	colorMode string
}

var (
	errSettingNotRuntimeTunable = errors.New("setting is not runtime tunable")

	logger *log.Logger //nolint:gochecknoglobals
	stdout io.Writer   //nolint:gochecknoglobals
	opts   options     //nolint:gochecknoglobals
	color  bool        //nolint:gochecknoglobals

	rootCmd = &cobra.Command{ //nolint:gochecknoglobals
		Use:   "cunicu",
		Short: "cunīcu is a user-space daemon managing WireGuard® interfaces to establish peer-to-peer connections in harsh network environments.",
		Long: `cunīcu is a user-space daemon managing WireGuard® interfaces to
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
	rootCmd.SetUsageTemplate(usageTemplate)

	cobra.OnInitialize(func() {
		if err := setupLogging(nil); err != nil {
			panic(err)
		}
	})

	f := rootCmd.Flags()
	f.SortFlags = false

	pf := rootCmd.PersistentFlags()
	pf.StringArrayVarP(&opts.logRules, "log-level", "d", nil, "log level filter rule (one of: debug, info, warn, error, dpanic, panic, and fatal) (default \"info\")")
	pf.StringArrayVarP(&opts.logFiles, "log-file", "l", nil, "path of a file to write logs to")
	pf.StringVarP(&opts.colorMode, "color", "q", "auto", "Enable colorization of output (one of: auto, always, never)")

	if err := rootCmd.RegisterFlagCompletionFunc("log-level", cobra.FixedCompletions([]string{"debug", "info", "warn", "error", "dpanic", "panic", "fatal"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}

	if err := rootCmd.RegisterFlagCompletionFunc("color", cobra.FixedCompletions([]string{"auto", "always", "never"}, cobra.ShellCompDirectiveNoFileComp)); err != nil {
		panic(err)
	}
}

type logConfigChangedHandler struct{}

func (h *logConfigChangedHandler) OnConfigChanged(key string, _, newValue any) error {
	switch key {
	case "log.rules":
		var newRules []string
		switch value := newValue.(type) {
		case string:
			newRules = []string{value}
		case []string:
			newRules = value
		}

		newRules = append(newRules, opts.logRules...)

		if len(newRules) == 0 {
			newRules = []string{"info"}
		}

		filter, err := log.ParseFilter(newRules)
		if err != nil {
			logger.Error("Failed to parse filter rules", zap.Error(err))
		}

		log.UpdateFilter(filter)

		logger.Debug("Updated log filter", zap.Strings("rules", newRules))

	default:
		return fmt.Errorf("%w: %s", errSettingNotRuntimeTunable, key)
	}

	return nil
}

func setupLogging(cfg *config.Config) error {
	// Color
	colorMode := opts.colorMode
	if cfg != nil && cfg.Log.Color != "" {
		colorMode = cfg.Log.Color
	}
	switch colorMode {
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

	// Files
	outputPaths := append([]string{"stdout"}, opts.logFiles...)

	if cfg != nil && cfg.Log.File != "" {
		outputPaths = append(outputPaths, cfg.Log.File)
	}

	// Rules
	rules := []string{}
	rules = append(rules, opts.logRules...)

	if cfg != nil {
		rules = append(rules, cfg.Log.Rules...)
		if cfg.Log.Level != "" {
			rules = append(rules, cfg.Log.Level)
		}
	}

	if len(rules) == 0 {
		rules = []string{"info"}
	}

	filterRule, err := log.ParseFilter(rules)
	if err != nil {
		return fmt.Errorf("failed to parse filter rules: %w", err)
	}

	logger, err = log.SetupLogging(filterRule, outputPaths, color)
	if err != nil {
		return err
	}

	if cfg != nil {
		h := &logConfigChangedHandler{}
		cfg.Meta.AddChangedHandler("log", h)
	}

	return nil
}

func main() {
	var execute func() error
	switch os.Args[0] {
	case "wg":
		execute = wgCmd.Execute
	default:
		execute = rootCmd.Execute
	}

	if err := execute(); err != nil {
		os.Exit(1)
	}
}
