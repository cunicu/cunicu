//go:build !linux

package watcher

func (w *Watcher) watchKernel() error {
	return errNotSupported
}
