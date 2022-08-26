//go:build unix

package util

import (
	"os"

	"golang.org/x/sys/unix"
)

const ReexecSelfSupported = true

func ReexecSelf() error {
	return unix.Exec("/proc/self/exe", os.Args, os.Environ())
}
