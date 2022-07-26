//go:build !linux

package proxy

import (
	"net"

	"riasc.eu/wice/pkg/errors"
)

type NAT struct{}

func NewNAT(ident string) (*NAT, error) {
	return nil, errors.ErrNotSupported
}

func (n *NAT) MasqueradeSourcePort(fromPort, toPort int, dest *net.UDPAddr) error {
	return errors.ErrNotSupported
}

func (n *NAT) RedirectNonSTUN(origPort, newPort int) error {
	return errors.ErrNotSupported
}

func (N *NAT) Close() error {
	return errors.ErrNotSupported
}
