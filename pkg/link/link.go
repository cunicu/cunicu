// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package link

import (
	"errors"
	"net"
)

const (
	EthernetMTU = 1500
)

const (
	TypeWireGuard = "wireguard"
)

var errNotSupported = errors.New("not supported") //nolint:unused

type Link interface {
	Close() error

	// Getter

	Name() string
	Index() int
	MTU() int
	Flags() net.Flags
	Type() string

	// Setter

	SetMTU(mtu int) error
	SetUp() error
	SetDown() error

	AddAddress(ip net.IPNet) error
	AddRoute(dst net.IPNet, gw net.IP, table int) error

	DeleteAddress(ip net.IPNet) error
	DeleteRoute(dst net.IPNet, table int) error
}
