//go:build !linux

package watcher

import (
	"riasc.eu/wice/pkg/errors"
)

func (w *Watcher) watchKernel() error {
	return errors.ErrNotSupported
}
