package selfupdate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/util"
)

// DownloadAndVerifyRelease downloads a released version of
// ɯice and saves it to target. It returns the version string for the newest
// version. The function printf is used to print progress information.
func DownloadAndVerifyRelease(ctx context.Context, rel *Release, target string, logger *zap.Logger) error {
	fn, sha256sums, err := getGithubDataFile(ctx, rel.Assets, checksumsFile)
	if err != nil {
		return err
	}

	logger.Info("Downloaded", zap.String("filename", fn))

	fn, sig, err := getGithubDataFile(ctx, rel.Assets, checksumsSigFile)
	if err != nil {
		return err
	}

	logger.Info("Downloaded", zap.String("filename", fn))

	if ok, err := GPGVerify(sha256sums, sig); err != nil {
		return fmt.Errorf("GPG signature verification of %s failed: %w", checksumsSigFile, err)
	} else if !ok {
		return fmt.Errorf("GPG signature verification of %s failed", checksumsSigFile)
	}

	logger.Info("GPG signature verification succeeded")

	ext := "gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}

	suffix := fmt.Sprintf("%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)
	downloadFilename, buf, err := getGithubDataFile(ctx, rel.Assets, suffix)
	if err != nil {
		return err
	}

	logger.Info("Downloaded", zap.String("filename", downloadFilename))

	wantHash, err := findHash(sha256sums, downloadFilename)
	if err != nil {
		return err
	}

	gotHash := sha256.Sum256(buf)
	if !bytes.Equal(wantHash, gotHash[:]) {
		return fmt.Errorf("checksum mismatch, want hash %02x, got %02x", wantHash, gotHash)
	}

	logger.Info("Checksum verification succeeded")

	var n int64
	if n, err = extractToFile(buf, downloadFilename, target); err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	logger.Info("Extraction succeeded", zap.String("size", util.PrettyBytes(n, false)))

	return nil
}
