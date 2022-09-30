package main

// derived from http://github.com/restic/restic

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/selfupdate"
	"go.uber.org/zap"
)

var (
	output string

	selfUpdateCmd = &cobra.Command{
		Use:   "selfupdate",
		Short: "Update the cunīcu binary",
		Long: `Downloads the latest stable release of cunīcu from GitHub and replaces the currently running binary.
After download, the authenticity of the binary is verified using the GPG signature on the release files.`,
		Run: selfUpdate,
	}
)

func init() {
	rootCmd.AddCommand(selfUpdateCmd)

	selfPath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	self := filepath.Base(selfPath)
	if strings.Contains(selfPath, "go-build") {
		self = "cunicu"
	}

	flags := selfUpdateCmd.Flags()
	flags.StringVarP(&output, "output", "o", self, "Save the downloaded file as `filename`")
}

func selfUpdate(cmd *cobra.Command, args []string) {
	logger := logger.Named("self-update")

	rel, err := selfupdate.SelfUpdate(output, logger)
	if err != nil {
		logger.Fatal("Self-update failed", zap.Error(err))
	}

	logger.Info("Successfully updated cunicu",
		zap.String("version", rel.Version),
		zap.String("filename", output),
	)
}
