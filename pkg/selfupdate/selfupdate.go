// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package selfupdate implements a cryptographically secured self-update mechanism which fetches binaries via GitHub's API.
package selfupdate

// derived from http://github.com/restic/restic

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/buildinfo"
	"github.com/stv0g/cunicu/pkg/log"
)

const (
	githubUser       = "stv0g"
	githubRepo       = "cunicu"
	binaryFile       = "cunicu"
	checksumsFile    = "checksums.txt"
	checksumsSigFile = checksumsFile + ".asc"
)

var (
	errOutputIsNotDir        = errors.New("output parent path is not a directory")
	errOutputIsNotNormalFile = errors.New("output path is not a normal file")
	errHashNotFound          = errors.New("hash not found")
	errArchiveMultipleFiles  = errors.New("archive contains more than one file")
)

func SelfUpdate(output string, logger *log.Logger) (*Release, error) {
	fi, err := os.Lstat(output)
	if err != nil {
		dirname := filepath.Dir(output)
		di, err := os.Lstat(dirname)
		if err != nil {
			return nil, fmt.Errorf("failed to stat: %w", err)
		}
		if !di.Mode().IsDir() {
			return nil, errOutputIsNotDir
		}
	} else if !fi.Mode().IsRegular() {
		return nil, errOutputIsNotNormalFile
	}

	curVersion := strings.TrimPrefix(buildinfo.Version, "v")

	logger.Info("Current version", zap.String("version", curVersion))

	rel, err := GitHubLatestRelease(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get latest release from GitHub: %w", err)
	}

	logger.Info("Latest version", zap.String("version", rel.Version))

	// We do a lexicographic comparison here to compare the semver versions.
	switch {
	case rel.Version == curVersion:
		logger.Info("Your cunicu version is up to date. Nothing to update.")
		return rel, nil
	case rel.Version < curVersion:
		logger.Warn("You are running an unreleased version of cunicu. Nothing to update.")
		return rel, nil
	default:
		logger.Info("Your cunicu version is outdated. Updating now!")
	}

	if err := DownloadAndVerifyRelease(context.Background(), rel, output, logger); err != nil {
		return rel, fmt.Errorf("failed to update cunicu: %w", err)
	}

	if err := VersionVerify(output, rel.Version); err != nil {
		return rel, fmt.Errorf("failed to update cunicu: %w", err)
	}

	return rel, nil
}

func findHash(buf []byte, filename string) (hash []byte, err error) {
	sc := bufio.NewScanner(bytes.NewReader(buf))
	for sc.Scan() {
		data := strings.Split(sc.Text(), "  ")
		if len(data) != 2 {
			continue
		}

		if data[1] == filename {
			h, err := hex.DecodeString(data[0])
			if err != nil {
				return nil, err
			}

			return h, nil
		}
	}

	return nil, fmt.Errorf("%w for: %s", errHashNotFound, filename)
}

func extractToFile(buf []byte, filename, target string) (int64, error) {
	mode := os.FileMode(0o755)

	// get information about the target file
	fi, err := os.Lstat(target)
	if err == nil {
		mode = fi.Mode()
	}

	ext := filepath.Ext(filename)

	var rd io.Reader = bytes.NewReader(buf)
	switch ext {
	case ".bz2":
		rd = bzip2.NewReader(rd)
	case ".gz":
		if rd, err = gzip.NewReader(rd); err != nil {
			return -1, err
		}
	}

	// Check if there is an archive
	ext = filepath.Ext(filename[0 : len(filename)-len(ext)])
	switch ext {
	case ".tar":
		trd := tar.NewReader(rd)
		rd = nil
		for {
			if hdr, err := trd.Next(); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return -1, fmt.Errorf("failed to open tar archive: %w", err)
			} else if hdr.Name == binaryFile {
				rd = trd
				break
			}
		}
		if rd == nil {
			return -1, fmt.Errorf("%w: %s", syscall.ENOENT, binaryFile)
		}
	case ".zip":
		zrd, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
		if err != nil {
			return -1, err
		}

		if len(zrd.File) != 1 {
			return -1, errArchiveMultipleFiles
		}

		file, err := zrd.File[0].Open()
		if err != nil {
			return -1, err
		}

		defer func() {
			_ = file.Close()
		}()

		rd = file
	}

	// Delete old file
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return -1, fmt.Errorf("failed to remove target file: %w", err)
	}

	dest, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return -1, err
	}

	n, err := io.Copy(dest, rd) //nolint:gosec
	if err != nil {
		_ = dest.Close()
		_ = os.Remove(dest.Name())
		return -1, fmt.Errorf("failed to copy: %w", err)
	}

	if err = dest.Close(); err != nil {
		return -1, fmt.Errorf("failed to close file: %w", err)
	}

	return n, nil
}
