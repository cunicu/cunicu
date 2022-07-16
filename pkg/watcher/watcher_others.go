//go:build !linux

package watcher

func WatchWireguardKernelInterfaces(chan InterfaceEvent, chan error) error {
	return nil
}
