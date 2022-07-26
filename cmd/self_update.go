package main

// derived from http://github.com/restic/restic

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/selfupdate"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update [flags]",
	Short: "Update the wice binary",
	Long: `
The command "self-update" downloads the latest stable release of wice from
GitHub and replaces the currently running binary. After download, the
authenticity of the binary is verified using the GPG signature on the release
files.
`,
	DisableAutoGenTag: true,
	Run:               selfUpdate,
}

// SelfUpdateOptions collects all options for the self-update command.
type SelfUpdateOptions struct {
	Output string
}

var selfUpdateOptions SelfUpdateOptions

func init() {
	RootCmd.AddCommand(selfUpdateCmd)

	flags := selfUpdateCmd.Flags()
	flags.StringVar(&selfUpdateOptions.Output, "output", "", "Save the downloaded file as `filename` (default: running binary itself)")
}

func selfUpdate(cmd *cobra.Command, args []string) {
	logger := logger.Named("self-update")

	if selfUpdateOptions.Output == "" {
		file, err := os.Executable()
		if err != nil {
			logger.Error("Unable to find executable", zap.Error(err))
		}

		selfUpdateOptions.Output = file
	}

	fi, err := os.Lstat(selfUpdateOptions.Output)
	if err != nil {
		dirname := filepath.Dir(selfUpdateOptions.Output)
		di, err := os.Lstat(dirname)
		if err != nil {
			logger.Fatal("Failed to stat", zap.Error(err))
		}
		if !di.Mode().IsDir() {
			logger.Fatal("Output parent path is not a directory, use --output to specify a different file path", zap.String("path", dirname))
		}
	} else {
		if !fi.Mode().IsRegular() {
			logger.Fatal("Output path is not a normal file, use --output to specify a different file path", zap.String("path", selfUpdateOptions.Output))
		}
	}

	logger.Info("Writing wice", zap.String("output", selfUpdateOptions.Output))

	v, err := selfupdate.DownloadLatestStableRelease(context.Background(), version, selfUpdateOptions.Output, logger)
	if err != nil {
		logger.Fatal("Failed to update ɯice", zap.Error(err))
	}

	if v != version {
		logger.Info("Successfully updated wice", zap.String("version", v))
	}
}
