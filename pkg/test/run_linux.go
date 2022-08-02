package test

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	gont "github.com/stv0g/gont/pkg"
)

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
