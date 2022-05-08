//go:build !linux

package proxy

import (
	"net"

	"riasc.eu/wice/internal/errors"
)

type NAT struct{}

func NewNAT(ident string) (*NAT, error) {
	return nil, errors.ErrNotSupported
}

func (n *NAT) MasqueradeSourcePort(fromPort, toPort uint16, dest *net.UDPAddr) error {
	return errors.ErrNotSupported
}

func (n *NAT) RedirectNonSTUN(origPort, newPort uint16) error {
	return errors.ErrNotSupported
}

func (N *NAT) Close() error {
	return errors.ErrNotSupported
}
