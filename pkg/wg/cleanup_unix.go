// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build unix

package wg

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/stv0g/cunicu/pkg/log"
)

func CleanupUserSockets() error {
	logger := log.Global.Named("wg")

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

		if filepath.Ext(p) == ".sock" {
			if c, err := net.Dial("unix", p); err == nil {
				if err := c.Close(); err != nil {
					return fmt.Errorf("failed to close socket: %w", err)
				}
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
