package test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sync"

	gont "github.com/stv0g/gont/pkg"
)

var (
	// Singleton for compiled wice executable
	binary      string
	binaryMutex sync.Mutex
)

func FindBaseDir() (string, error) {
	p, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for p != "." {
		if _, err := os.Stat(path.Join(p, ".git")); err != nil {
			if os.IsNotExist(err) {
				p = filepath.Dir(p)
			} else {
				return "", err
			}
		} else {
			return p, nil
		}
	}

	return "", os.ErrNotExist
}

func BuildBinary() (string, error) {
	binaryMutex.Lock()
	defer binaryMutex.Unlock()

	if binary == "" {
		binary = "/tmp/wice"

		base, err := FindBaseDir()
		if err != nil {
			return "", fmt.Errorf("failed to find base dir: %w", err)
		}

		wd, _ := os.Getwd()
		os.Chdir(base)
		defer os.Chdir(wd)
		// zap.L().Info("Base dir", zap.String("dir", base), zap.String("wd", wd))

		cmd := exec.Command("go", "build", "-buildvcs=false", "-o", binary, "./cmd/wice/")
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to build wice: %w\n%s", err, out)
		}
	}

	return binary, nil
}

func RunWice(h *gont.Host, args ...interface{}) ([]byte, *exec.Cmd, error) {
	bin, err := BuildBinary()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build binary: %w", err)
	}

	return h.Run(bin, args)
}

func StartWice(h *gont.Host, args ...interface{}) (io.Reader, io.Reader, *exec.Cmd, error) {
	bin, err := BuildBinary()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to build binary: %w", err)
	}

	return h.Start(bin, args...)
}
