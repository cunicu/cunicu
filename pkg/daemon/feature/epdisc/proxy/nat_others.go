//go:build !linux

package proxy

import (
	"errors"
	"net"
)

var errNotSupported = errors.New("not supported")

type (
	NATRule struct{}
	NAT     struct{}
)

func NewNAT(ident string) (*NAT, error) {
	return nil, errNotSupported
}

func (n *NAT) MasqueradeSourcePort(fromPort, toPort int, dest *net.UDPAddr) (*NATRule, error) {
	return nil, errNotSupported
}

func (n *NAT) RedirectNonSTUN(origPort, newPort int) (*NATRule, error) {
	return nil, errNotSupported
}

func (n *NAT) Close() error {
	return errNotSupported
}

func (nr *NATRule) Delete() error {
	return errNotSupported
}
