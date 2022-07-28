//go:build !linux

package wg

func WireGuardModuleExists() bool {
	return false
}
