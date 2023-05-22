//go:build unix

package os

import (
	"os"

	"golang.org/x/sys/unix"
)

const ReexecSelfSupported = true

func ReexecSelf() error {
	return unix.Exec("/proc/self/exe", os.Args, os.Environ())
}
