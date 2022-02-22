package e2e

import (
	"fmt"

	g "github.com/stv0g/gont/pkg"
)

var (
	// Singleton for compiled wice executable
	wiceBinary string
)

func buildWICE(n *g.Network) (string, error) {
	if wiceBinary != "" {
		return wiceBinary, nil
	}

	wiceBinary := "/tmp/wice"

	if out, _, err := n.HostNode.Run("go", "build", "-o", wiceBinary, "../../cmd/wice/"); err != nil {
		return "", fmt.Errorf("failed to build wice: %w\n%s", err, out)
	}

	return wiceBinary, nil
}
