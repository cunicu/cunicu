//go:build unix

package wg

import (
	"io/fs"
	"net"
	"path"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

func CleanupUserSockets() error {
	logger := zap.L().Named("wg")

	return filepath.Walk(SocketPath, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !i.IsDir() && path.Ext(p) == ".sock" {
			if c, err := net.Dial("unix", p); err == nil {
				c.Close()
			} else {
				logger.Warn("Delete stale WireGuard user socket", zap.String("path", p))

				return unix.Unlink(p)
			}
		}

		return nil
	})
}
