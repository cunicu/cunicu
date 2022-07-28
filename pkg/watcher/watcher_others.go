//go:build !linux

package watcher

func WatchWireGuardKernelInterfaces(chan InterfaceEvent, chan error) error {
	return nil
}
