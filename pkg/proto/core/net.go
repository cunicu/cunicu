// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package core

import "net"

func Address(i net.IP) *IPAddress {
	if b := i.To4(); b != nil {
		i = b
	}

	return &IPAddress{
		Addr: i,
	}
}

func (i *IPAddress) Address() net.IP {
	return (net.IP)(i.Addr)
}

func Prefix(i net.IPNet) *IPPrefix {
	ones, _ := i.Mask.Size()

	return &IPPrefix{
		Addr:   i.IP,
		Pfxlen: uint32(ones),
	}
}

func (i *IPPrefix) Prefix() *net.IPNet {
	ones := int(i.Pfxlen)
	bits := len(i.Addr) * 8

	return &net.IPNet{
		IP:   i.Addr,
		Mask: net.CIDRMask(ones, bits),
	}
}
