//go:build !linux

package proxy

import (
	"net"

	"github.com/stv0g/cunicu/pkg/errors"
)

type NATRule struct{}
type NAT struct{}

func NewNAT(ident string) (*NAT, error) {
	return nil, errors.ErrNotSupported
}

func (n *NAT) MasqueradeSourcePort(fromPort, toPort int, dest *net.UDPAddr) (*NATRule, error) {
	return nil, errors.ErrNotSupported
}

func (n *NAT) RedirectNonSTUN(origPort, newPort int) (*NATRule, error) {
	return nil, errors.ErrNotSupported
}

func (n *NAT) Close() error {
	return errors.ErrNotSupported
}

func (nr *NATRule) Delete() error {
	return errors.ErrNotSupported
}
