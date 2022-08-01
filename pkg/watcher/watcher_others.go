//go:build !linux

package watcher

import (
	"riasc.eu/wice/pkg/errors"
)

func WatchWireGuardKernelInterfaces(chan InterfaceEvent, chan error) error {
	return nil
}

func (w *Watcher) watchKernel() error {
	return errors.ErrNotSupported
}
