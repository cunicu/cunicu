package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"
)

var (
	outputDir string

	docsCmd = &cobra.Command{
		Use:    "docs",
		Short:  "Generate documentation for the cunīcu commands",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := docsMarkdown(cmd, args); err != nil {
				return err
			}

			return docsManpage(cmd, args)
		},
	}

	docsMarkdownCmd = &cobra.Command{
		Use:   "markdown",
		Short: "Generate markdown docs",
		RunE:  docsMarkdown,
		Args:  cobra.NoArgs,
	}

	docsManpageCmd = &cobra.Command{
		Use:   "man",
		Short: "Generate manpages",
		RunE:  docsManpage,
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(docsCmd)

	docsCmd.AddCommand(docsManpageCmd)
	docsCmd.AddCommand(docsMarkdownCmd)

	pf := docsCmd.PersistentFlags()
	pf.StringVar(&outputDir, "output-dir", "./docs/usage", "Output directory of generated documentation")
}

func docsMarkdown(cmd *cobra.Command, args []string) error {
	dir := filepath.Join(outputDir, "md")

	//#nosec G301 -- Doc directories must be world readable
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return doc.GenMarkdownTree(rootCmd, dir)
}

func docsManpage(cmd *cobra.Command, args []string) error {
	dir := filepath.Join(outputDir, "man")

	//#nosec G301 -- Doc directories must be world readable
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	header := &doc.GenManHeader{
		Title:   "cunicu",
		Section: "3",
		Source:  "https://github.com/stv0g/cunicu",
		Date:    buildinfo.BuiltDate,
	}

	return doc.GenManTree(rootCmd, header, dir)
}
