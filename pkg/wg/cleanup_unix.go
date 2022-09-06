//go:build unix

package wg

import (
	"errors"
	"net"
	"os"
	"path"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func CleanupUserSockets() error {
	logger := zap.L().Named("wg")

	// Ignore non-existing dir
	if _, err := os.Stat(SocketPath); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	des, err := os.ReadDir(SocketPath)
	if err != nil {
		return err
	}

	for _, de := range des {
		p := filepath.Join(SocketPath, de.Name())

		if path.Ext(p) == ".sock" {
			if c, err := net.Dial("unix", p); err == nil {
				c.Close()
			} else if !errors.Is(err, unix.ENOENT) {
				logger.Warn("Delete stale WireGuard user socket", zap.String("path", p))

				if err := unix.Unlink(p); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
