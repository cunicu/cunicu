//go:build unix

package wg

import (
	"errors"
	"io/fs"
	"net"
	"os"
	"path"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func CleanupUserSockets() error {
	logger := zap.L().Named("wg")

	if _, err := os.Stat(SocketPath); err != nil && errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return filepath.Walk(SocketPath, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !i.IsDir() && path.Ext(p) == ".sock" {
			if c, err := net.Dial("unix", p); err == nil {
				c.Close()
			} else if !errors.Is(err, unix.ENOENT) {
				logger.Warn("Delete stale WireGuard user socket", zap.String("path", p))

				return unix.Unlink(p)
			}
		}

		return nil
	})
}
