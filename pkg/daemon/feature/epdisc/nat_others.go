// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package epdisc

import (
	"net"
)

type (
	NATRule struct{}
	NAT     struct{}
)

func NewNAT(_ string) (*NAT, error) {
	return nil, errNotSupported
}

func (n *NAT) MasqueradeSourcePort(_, _ int, _ *net.UDPAddr) (*NATRule, error) {
	return nil, errNotSupported
}

func (n *NAT) RedirectNonSTUN(_, _ int) (*NATRule, error) {
	return nil, errNotSupported
}

func (n *NAT) Close() error {
	return errNotSupported
}

func (nr *NATRule) Delete() error {
	return errNotSupported
}
