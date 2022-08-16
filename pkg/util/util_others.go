//go:build !windows

package util

import (
	"golang.org/x/sys/unix"
)

const (
	SigUpdate = unix.SIGUSR1
)
