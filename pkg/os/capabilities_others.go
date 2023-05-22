//go:build !(linux || darwin || freebsd || openbsd || netbsd || plan9 || dragonfly || solaris)

package os

func HasAdminPrivileges() bool {
	return false
}
