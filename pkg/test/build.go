package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	// Singleton for compiled ɯice executable
	binary      string
	binaryMutex sync.Mutex

	packageName = "riasc.eu/wice"
)

func FindBaseDir() (string, error) {
	p, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for p != "." {
		if _, err := os.Stat(filepath.Join(p, ".git")); err != nil {
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

func BuildBinary(coverage bool) (string, error) {
	binaryMutex.Lock()
	defer binaryMutex.Unlock()

	if binary == "" {
		binaryDir, err := os.MkdirTemp("", "wice-build-*")
		if err != nil {
			return "", err
		}
		binary = filepath.Join(binaryDir, "wice")

		base, err := FindBaseDir()
		if err != nil {
			return "", fmt.Errorf("failed to find base dir: %w", err)
		}

		var pkg = filepath.Join(base, "cmd")
		var cmd *exec.Cmd
		if coverage {
			//#nosec G204 -- Just for testing
			cmd = exec.Command("go", "test", "-o", binary, "-buildvcs=false", "-cover", "-covermode=count", "-coverpkg="+packageName+"/...", "-c", "-tags", "testmain", pkg)
		} else {
			//#nosec G204 -- Just for testing
			cmd = exec.Command("go", "build", "-buildvcs=false", "-o", binary, pkg)
		}

		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to build ɯice: %w\n%s", err, out)
		}
	}

	return binary, nil
}
