//go:build !linux

package daemon

func (w *Watcher) watchKernelInterfaces() error {
	return errNotSupported
}
