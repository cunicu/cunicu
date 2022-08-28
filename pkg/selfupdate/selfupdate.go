// Package selfupdate implements a cryptographically secured self-update mechanism which fetches binaries via GitHub's API.
package selfupdate

// derived from http://github.com/restic/restic

import (
	"archive/zip"
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	githubUser = "stv0g"
	githubRepo = "wice"

	checksumsFile    = "checksums.txt"
	checksumsSigFile = checksumsFile + ".asc"
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

func extractToFile(buf []byte, filename, target string) (int64, error) {
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
	case ".gz":
		if rd, err = gzip.NewReader(rd); err != nil {
			return -1, err
		}
	case ".zip":
		zrd, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
		if err != nil {
			return -1, err
		}

		if len(zrd.File) != 1 {
			return -1, fmt.Errorf("ZIP archive contains more than one file")
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

	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return -1, fmt.Errorf("failed to remove target file: %v", err)
	}

	//#nosec G304 -- No file inclusion possible as we are writing only.
	dest, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return -1, err
	}

	//#nosec G110 -- We only download from safe locations (GitHub releases)
	n, err := io.Copy(dest, rd)
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
