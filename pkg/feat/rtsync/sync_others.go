//go:build !linux

package rtsync

import (
	"github.com/stv0g/cunicu/pkg/core"
	"github.com/stv0g/cunicu/pkg/errors"
)

func (s *RouteSync) removeKernel(p *core.Peer) error {
	return errors.ErrNotSupported
}

func (s *RouteSync) syncKernel() error {
	return errors.ErrNotSupported
}

func (s *RouteSync) watchKernel() {}
