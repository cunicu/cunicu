//go:build !(linux || darwin || freebsd || openbsd || netbsd || plan9 || dragonfly || solaris)

package util

func HasAdminPrivileges() bool {
	return false
}
