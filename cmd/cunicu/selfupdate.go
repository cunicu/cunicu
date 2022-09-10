package main

// derived from http://github.com/restic/restic

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stv0g/cunicu/pkg/selfupdate"
	"github.com/stv0g/cunicu/pkg/util/buildinfo"
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

	file, err := os.Executable()
	if err != nil {
		panic(err)
	}

	if strings.Contains(file, "go-build") {
		file = "cunicu"
	}

	flags := selfUpdateCmd.Flags()
	flags.StringVarP(&output, "output", "o", file, "Save the downloaded file as `filename`")
}

func selfUpdate(cmd *cobra.Command, args []string) {
	logger := logger.Named("self-update")

	fi, err := os.Lstat(output)
	if err != nil {
		dirname := filepath.Dir(output)
		di, err := os.Lstat(dirname)
		if err != nil {
			logger.Fatal("Failed to stat", zap.Error(err))
		}
		if !di.Mode().IsDir() {
			logger.Fatal("Output parent path is not a directory, use --output to specify a different file path", zap.String("path", dirname))
		}
	} else {
		if !fi.Mode().IsRegular() {
			logger.Fatal("Output path is not a normal file, use --output to specify a different file path", zap.String("path", output))
		}
	}

	curVersion := strings.TrimPrefix(buildinfo.Version, "v")

	logger.Info("Current version", zap.String("version", curVersion))

	rel, err := selfupdate.GitHubLatestRelease(context.Background())
	if err != nil {
		logger.Fatal("Failed to get latest release from GitHub", zap.Error(err))
	}

	logger.Info("Latest version", zap.String("version", rel.Version))

	// We do a lexicographic comparison here to compare the semver versions.
	if rel.Version == curVersion {
		logger.Info("Your cunicu version is up to date. Nothing to update.")
		return
	} else if rel.Version < curVersion {
		logger.Warn("You are running an unreleased version of cunicu. Nothing to update.")
		return
	} else {
		logger.Info("Your cunicu version is outdated. Updating now!")
	}

	if err := selfupdate.DownloadAndVerifyRelease(context.Background(), rel, output, logger); err != nil {
		logger.Fatal("Failed to update cunicu", zap.Error(err))
	}

	if err := selfupdate.VersionVerify(output, rel.Version); err != nil {
		logger.Fatal("Failed to update cunicu", zap.Error(err))
	}

	logger.Info("Successfully updated cunicu",
		zap.String("version", rel.Version),
		zap.String("filename", output),
	)
}
