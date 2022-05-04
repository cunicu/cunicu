package e2e

import (
	"fmt"
	"sync"

	g "github.com/stv0g/gont/pkg"
)

var (
	// Singleton for compiled wice executable
	binary      string
	binaryMutex sync.Mutex
)

func buildBinary(n *g.Network) (string, error) {
	binaryMutex.Lock()
	defer binaryMutex.Unlock()

	if binary != "" {
		return binary, nil
	}

	wiceBinary := "/tmp/wice"

	if out, _, err := n.HostNode.Run("go", "build", "-o", wiceBinary, "../../cmd/wice/"); err != nil {
		return "", fmt.Errorf("failed to build wice: %w\n%s", err, out)
	}

	return wiceBinary, nil
}
