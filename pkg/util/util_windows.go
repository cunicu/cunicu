//go:build windows

package util

import (
	"syscall"
)

const (
	SigUpdate = syscall.Signal(-1) // not supported
)
