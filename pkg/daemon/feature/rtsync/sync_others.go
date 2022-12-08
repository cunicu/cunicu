//go:build !linux

package rtsync

import (
	"github.com/stv0g/cunicu/pkg/core"
)

func (i *Interface) removeKernel(p *core.Peer) error {
	return errNotSupported
}

func (i *Interface) syncKernel() error {
	return errNotSupported
}

func (i *Interface) watchKernel() error {
	return errNotSupported
}
