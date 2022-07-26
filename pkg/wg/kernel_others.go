//go:build !linux

package wg

func WireguardModuleExists() bool {
	return false
}
