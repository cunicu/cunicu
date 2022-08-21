package cmd

// derived from http://github.com/restic/restic

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"riasc.eu/wice/pkg/selfupdate"
)

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update [flags]",
	Short: "Update the ɯice binary",
	Long: `
The command "self-update" downloads the latest stable release of ɯice from
GitHub and replaces the currently running binary. After download, the
authenticity of the binary is verified using the GPG signature on the release
files.
`,
	Run: selfUpdate,
}

// SelfUpdateOptions collects all options for the self-update command.
type SelfUpdateOptions struct {
	Output string
}

var selfUpdateOptions SelfUpdateOptions

func init() {
	RootCmd.AddCommand(selfUpdateCmd)

	file, err := os.Executable()
	if err != nil {
		panic(err)
	}

	flags := selfUpdateCmd.Flags()
	flags.StringVarP(&selfUpdateOptions.Output, "output", "o", file, "Save the downloaded file as `filename`")
}

func selfUpdate(cmd *cobra.Command, args []string) {
	logger := logger.Named("self-update")

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

	curVersion := strings.TrimPrefix(version, "v")

	logger.Info("Current version", zap.String("version", curVersion))

	rel, err := selfupdate.GitHubLatestRelease(context.Background())
	if err != nil {
		logger.Fatal("Failed to get latest release from GitHub", zap.Error(err))
	}

	logger.Info("Latest version", zap.String("version", rel.Version))

	// We do a lexicographic comparison here to compare the
	// semver versions.
	if rel.Version == curVersion {
		logger.Info("Your ɯice version is up to date. Aborting...")
		return
	} else if rel.Version < curVersion {
		logger.Warn("You are running an unreleased version of ɯice. Aborting...")
		return
	} else {
		logger.Info("Your ɯice version is out dated. Updating...")
	}

	if err := selfupdate.DownloadAndVerifyRelease(context.Background(), rel, selfUpdateOptions.Output, logger); err != nil {
		logger.Fatal("Failed to update ɯice", zap.Error(err))
	}

	logger.Info("Successfully updated ɯice",
		zap.String("version", rel.Version),
		zap.String("filename", selfUpdateOptions.Output),
	)
}
