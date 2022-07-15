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

	packageName = "riasc.eu/wice"
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

func BuildBinary(coverage bool) (string, error) {
	binaryMutex.Lock()
	defer binaryMutex.Unlock()

	if binary == "" {
		binary = TempFileName("wice-", "")

		base, err := FindBaseDir()
		if err != nil {
			return "", fmt.Errorf("failed to find base dir: %w", err)
		}

		var pkg = filepath.Join(base, "cmd/wice")
		var cmd *exec.Cmd
		if coverage {
			//#nosec G204 -- Just for testing
			cmd = exec.Command("go", "test", "-o", binary, "-buildvcs=false", "-cover", "-covermode=count", "-coverpkg="+packageName+"/...", "-c", "-tags", "testmain", pkg)
		} else {
			//#nosec G204 -- Just for testing
			cmd = exec.Command("go", "build", "-buildvcs=false", "-o", binary, pkg)
		}

		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to build wice: %w\n%s", err, out)
		}
	}

	return binary, nil
}

func RunWice(h *gont.Host, args ...any) ([]byte, *exec.Cmd, error) {
	bin, err := BuildBinary(false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build binary: %w", err)
	}

	return h.Run(bin, args...)
}

func StartWice(h *gont.Host, args ...any) (io.Reader, io.Reader, *exec.Cmd, error) {
	bin, err := BuildBinary(false)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to build binary: %w", err)
	}

	return h.Start(bin, args...)
}

func StartWiceWithCoverage(h *gont.Host, args ...any) (io.Reader, io.Reader, *exec.Cmd, error) {
	bin, err := BuildBinary(true)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to build binary: %w", err)
	}

	base, err := FindBaseDir()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to find base dir: %w", err)
	}

	covPath := fmt.Sprintf("coverprofile-%s.out", h.Name())
	covPath = filepath.Join(base, covPath)
	newArgs := []any{"-test.run=^TestMain$", "-test.coverprofile=" + covPath, "--"}
	newArgs = append(newArgs, args...)

	return h.Start(bin, newArgs...)
}
