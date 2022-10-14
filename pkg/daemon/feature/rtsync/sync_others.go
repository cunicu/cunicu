//go:build !linux

package rtsync

import (
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/errors"
)

func (s *Interface) removeKernel(p *core.Peer) error {
	return errors.ErrNotSupported
}

func (s *Interface) syncKernel() error {
	return errors.ErrNotSupported
}

func (s *Interface) watchKernel() error {
	return errors.ErrNotSupported
}
