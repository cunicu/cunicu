package test

import (
	"fmt"
	"os/exec"
	"sync"
)

var (
	// Singleton for compiled wice executable
	binary      string
	binaryMutex sync.Mutex
)

func BuildBinary() (string, error) {
	binaryMutex.Lock()
	defer binaryMutex.Unlock()

	if binary == "" {
		binary = "/tmp/wice"

		cmd := exec.Command("go", "build", "-o", binary, "../../cmd/wice/")
		if out, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to build wice: %w\n%s", err, out)
		}
	}

	return binary, nil
}
