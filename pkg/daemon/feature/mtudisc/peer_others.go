//go:build !linux

package mtudisc

import (
	xerrors "github.com/stv0g/cunicu/errors"
)

func (p *Peer) DiscoverPathMTU() (int, error) {
	return -1, xerrors.ErrNotSupported
}
