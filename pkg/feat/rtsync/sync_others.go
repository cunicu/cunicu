//go:build !linux

package rtsync

import (
	"riasc.eu/wice/pkg/core"
	"riasc.eu/wice/pkg/errors"
)

func (s *RouteSync) removeKernel(p *core.Peer) error {
	return errors.ErrNotSupported
}

func (s *RouteSync) syncKernel() error {
	return errors.ErrNotSupported
}

func (s *RouteSync) watchKernel() {}
