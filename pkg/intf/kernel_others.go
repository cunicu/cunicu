//go:build !linux

package intf

func WireguardModuleExists() bool {
	return false
}
