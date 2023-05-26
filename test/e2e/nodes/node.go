// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nodes

import (
	"fmt"
	"time"

	g "github.com/stv0g/gont/v2/pkg"
	"golang.org/x/sys/unix"
)

const (
	KillTimeout = 30 * time.Second
)

type Node interface {
	g.Node

	Start(binary, dir string, args ...any) error
	Stop() error
	Close() error
}

func GracefullyTerminate(cmd *g.Cmd) error {
	if err := cmd.Process.Signal(unix.SIGTERM); err != nil {
		return err
	}

	// Forcefully kill agent if it did not terminate after 10secs
	timer := time.AfterFunc(KillTimeout, func() {
		if err := cmd.Process.Kill(); err != nil {
			panic(fmt.Errorf("failed to kill process: %w", err))
		}
	})
	defer timer.Stop()

	return cmd.Wait()
}

type IncrementingDebugPort struct {
	Port int
}
