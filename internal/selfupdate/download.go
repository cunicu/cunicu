package selfupdate

// derived from http://github.com/restic/restic

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/bzip2"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

const (
	githubUser = "stv0g"
	githubRepo = "wice"
)

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

	return nil, fmt.Errorf("hash for file %v not found", filename)
}

func extractToFile(buf []byte, filename, target string, logger *zap.Logger) error {
	var mode = os.FileMode(0755)

	// get information about the target file
	fi, err := os.Lstat(target)
	if err == nil {
		mode = fi.Mode()
	}

	var rd io.Reader = bytes.NewReader(buf)
	switch filepath.Ext(filename) {
	case ".bz2":
		rd = bzip2.NewReader(rd)
	case ".zip":
		zrd, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
		if err != nil {
			return err
		}

		if len(zrd.File) != 1 {
			return fmt.Errorf("ZIP archive contains more than one file")
		}

		file, err := zrd.File[0].Open()
		if err != nil {
			return err
		}

		defer func() {
			_ = file.Close()
		}()

		rd = file
	}

	err = os.Remove(target)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("unable to remove target file: %v", err)
	}

	dest, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}

	n, err := io.Copy(dest, rd)
	if err != nil {
		_ = dest.Close()
		_ = os.Remove(dest.Name())
		return err
	}

	err = dest.Close()
	if err != nil {
		return err
	}

	logger.Sugar().Info("Saved %d bytes in %v", n, dest.Name())
	return nil
}

// DownloadLatestStableRelease downloads the latest stable released version of
// ɯice and saves it to target. It returns the version string for the newest
// version. The function printf is used to print progress information.
func DownloadLatestStableRelease(ctx context.Context, target, currentVersion string, logger *zap.Logger) (version string, err error) {
	logger.Info("Find latest release of ɯice at GitHub")

	rel, err := GitHubLatestRelease(ctx, githubUser, githubRepo)
	if err != nil {
		return "", err
	}

	if rel.Version == currentVersion {
		logger.Info("ɯice is up to date")
		return currentVersion, nil
	}

	logger.Sugar().Infof("Latest version is %v", rel.Version)

	_, sha256sums, err := getGithubDataFile(ctx, rel.Assets, "checksums.txt", logger)
	if err != nil {
		return "", err
	}

	_, sig, err := getGithubDataFile(ctx, rel.Assets, "checksums.txt.asc", logger)
	if err != nil {
		return "", err
	}

	ok, err := GPGVerify(sha256sums, sig)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", fmt.Errorf("GPG signature verification of the file SHA256SUMS failed")
	}

	logger.Info("GPG signature verification succeeded")

	ext := "bz2"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	suffix := fmt.Sprintf("%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	downloadFilename, buf, err := getGithubDataFile(ctx, rel.Assets, suffix, logger)
	if err != nil {
		return "", err
	}

	logger.Info("Downloaded", zap.String("file", downloadFilename))

	wantHash, err := findHash(sha256sums, downloadFilename)
	if err != nil {
		return "", err
	}

	gotHash := sha256.Sum256(buf)
	if !bytes.Equal(wantHash, gotHash[:]) {
		return "", fmt.Errorf("SHA256 hash mismatch, want hash %02x, got %02x", wantHash, gotHash)
	}

	err = extractToFile(buf, downloadFilename, target, logger)
	if err != nil {
		return "", err
	}

	return rel.Version, nil
}
