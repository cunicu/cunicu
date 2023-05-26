// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"net"
	"net/netip"

	"golang.zx2c4.com/wireguard/conn"
)

// Compile-time assertion
var _ conn.Endpoint = (*BindEndpoint)(nil)

type BindEndpoint struct {
	netip.AddrPort

	Conn BindConn
}

func (BindEndpoint) ClearSrc() {}

func (ep BindEndpoint) DstIP() netip.Addr {
	return ep.AddrPort.Addr()
}

func (ep BindEndpoint) DstToBytes() []byte {
	b, _ := ep.AddrPort.MarshalBinary()
	return b
}

func (ep BindEndpoint) DstToString() string {
	return ep.AddrPort.String()
}

func (ep BindEndpoint) SrcIP() netip.Addr {
	return netip.Addr{} // not supported
}

func (ep BindEndpoint) SrcToString() string {
	return "" // not supported
}

func (ep BindEndpoint) DstUDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   ep.AddrPort.Addr().AsSlice(),
		Port: int(ep.Port()),
	}
}
