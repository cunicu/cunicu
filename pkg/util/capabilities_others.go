//go:build !linux

package util

func HasAdminPrivileges() bool {
	return false // TODO
}
