//go:build !linux

package core

func WatchWireguardKernelInterfaces(chan InterfaceEvent, chan error) error {
	return nil
}
