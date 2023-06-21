// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/stv0g/cunicu/pkg/buildinfo"
)

type docsOptions struct {
	outputDir       string
	withFrontMatter bool
}

func init() { //nolint:gochecknoinits
	opts := &docsOptions{}
	cmd := &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for the cunīcu commands",
		Long:   `When used without a sub-command, both the Markdown documentation and Man-pages will be generated.`,
		Hidden: true,
		Args:   cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := docsMarkdown(cmd, args, opts); err != nil {
				logger.Fatal("Failed to generate markdown docs", zap.Error(err))
			}

			if err := docsManpage(cmd, args, opts); err != nil {
				logger.Fatal("Failed to generate Manpage docs", zap.Error(err))
			}
		},
	}

	docsMarkdownCmd := &cobra.Command{
		Use:   "markdown",
		Short: "Generate markdown docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return docsMarkdown(cmd, args, opts)
		},
		Args: cobra.NoArgs,
	}

	docsManpageCmd := &cobra.Command{
		Use:   "man",
		Short: "Generate manpages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return docsManpage(cmd, args, opts)
		},
		Args: cobra.NoArgs,
	}

	rootCmd.AddCommand(cmd)

	cmd.AddCommand(docsManpageCmd)
	cmd.AddCommand(docsMarkdownCmd)

	pf := cmd.PersistentFlags()
	pf.StringVar(&opts.outputDir, "output-dir", "./docs/usage", "Output directory of generated documentation")
	pf.BoolVar(&opts.withFrontMatter, "with-frontmatter", false, "Prepend a frontmatter to the generated Markdown files as used by our static website generator")
}

func docsMarkdown(_ *cobra.Command, _ []string, opts *docsOptions) error {
	dir := filepath.Join(opts.outputDir, "md")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filePrepender := func(path string) string {
		if !opts.withFrontMatter {
			return ""
		}

		filename := filepath.Base(path)
		filename = strings.TrimSuffix(filename, ".md")
		parts := strings.Split(filename, "_")

		fm := struct {
			Title            string   `yaml:"title"`
			SidebarLabel     string   `yaml:"sidebar_label,omitempty"`
			SidebarClassName string   `yaml:"sidebar_class_name"`
			Slug             string   `yaml:"slug"`
			HideTitle        bool     `yaml:"hide_title"`
			Keywords         []string `yaml:"keywords"`
		}{
			Title:            strings.Join(parts, " "),
			SidebarClassName: "command-name",
			Slug:             fmt.Sprintf("/usage/man/%s", strings.Join(parts[1:], "/")),
			HideTitle:        true,
			Keywords:         []string{"manpage"},
		}

		if len(parts) > 1 {
			fm.SidebarLabel = strings.Join(parts[1:], " ")
		}

		fmYaml, err := yaml.Marshal(fm)
		if err != nil {
			return ""
		}

		return fmt.Sprintf("---\n%s---\n\n", fmYaml)
	}

	// The linkHandler can be used to customize the rendered internal links to the commands, given a filename:
	linkHandler := func(name string) string {
		return name
	}

	return doc.GenMarkdownTreeCustom(rootCmd, dir, filePrepender, linkHandler)
}

func docsManpage(_ *cobra.Command, _ []string, opts *docsOptions) error {
	dir := filepath.Join(opts.outputDir, "man")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	header := &doc.GenManHeader{
		Title:  "cunīcu",
		Source: "https://github.com/stv0g/cunicu",
		Date:   buildinfo.Date,
	}

	return doc.GenManTree(rootCmd, header, dir)
}
