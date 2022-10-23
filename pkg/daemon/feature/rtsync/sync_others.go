//go:build !linux

package rtsync

import (
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/errors"
)

func (i *Interface) removeKernel(p *core.Peer) error {
	return errors.ErrNotSupported
}

func (i *Interface) syncKernel() error {
	return errors.ErrNotSupported
}

func (i *Interface) watchKernel() error {
	return errors.ErrNotSupported
}
