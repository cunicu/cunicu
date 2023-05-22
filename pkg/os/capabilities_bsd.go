//go:build darwin || freebsd || openbsd || netbsd || plan9 || dragonfly || solaris

package os

import (
	"os"
)

func HasAdminPrivileges() bool {
	return os.Geteuid() == 0
}
