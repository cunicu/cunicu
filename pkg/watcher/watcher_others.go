//go:build !linux

package watcher

import (
	"github.com/stv0g/cunicu/pkg/errors"
)

func (w *Watcher) watchKernel() error {
	return errors.ErrNotSupported
}
