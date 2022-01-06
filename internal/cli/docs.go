package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

type runEfunc func(cmd *cobra.Command, args []string) error

var (
	outputDir string = "./"
)

func NewDocsCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation for the wice commands",
		// Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := docsMarkdown(rootCmd); err != nil {
				return err
			}
			if err := docsManpage(rootCmd); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "markdown",
		Short: "Generate markdown docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			return docsMarkdown(rootCmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "man",
		Short: "Generate manpages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return docsManpage(rootCmd)
		},
	})

	pf := cmd.PersistentFlags()
	pf.StringVar(&outputDir, "output-dir", "./docs", "Output directory of generated documenation")

	return cmd
}

func docsMarkdown(rootCmd *cobra.Command) error {
	dir := filepath.Join(outputDir, "md")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := doc.GenMarkdownTree(rootCmd, dir); err != nil {
		log.Fatal(err)
	}

	return nil
}

func docsManpage(rootCmd *cobra.Command) error {
	dir := filepath.Join(outputDir, "man")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	header := &doc.GenManHeader{
		Title:   "MINE",
		Section: "3",
	}

	if err := doc.GenManTree(rootCmd, header, dir); err != nil {
		log.Fatal(err)
	}

	return nil
}
