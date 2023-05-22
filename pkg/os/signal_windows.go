//go:build windows

package os

import (
	"syscall"
)

const (
	SigUpdate = syscall.Signal(-1) // not supported
)
