//go:build !linux
// +build !linux

package intf

func WatchWireguardKernelInterfaces(chan InterfaceEvent, chan error) error {
	return nil
}
