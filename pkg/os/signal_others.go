//go:build !windows

package os

import (
	"golang.org/x/sys/unix"
)

const (
	SigUpdate = unix.SIGUSR1
)
