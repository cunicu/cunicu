// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"bytes"
	"net"
	"os"
	"time"
)

type packet struct {
	addr net.Addr
	buf  []byte
}

type PacketPipe struct {
	pkts chan packet

	readDeadline  <-chan time.Time
	writeDeadline <-chan time.Time

	addr net.Addr
}

func NewPacketPipe(lAddr net.Addr, depth int) *PacketPipe {
	pp := &PacketPipe{
		pkts: make(chan packet, depth),
		addr: lAddr,
	}

	return pp
}

func (p *PacketPipe) ReadFrom(buf []byte) (int, net.Addr, error) {
	select {
	case <-p.readDeadline:
		return -1, nil, os.ErrDeadlineExceeded
	case c := <-p.pkts:
		return copy(buf, c.buf), c.addr, nil
	}
}

func (p *PacketPipe) WriteFrom(buf []byte, addr net.Addr) (int, error) {
	select {
	case <-p.writeDeadline:
		return -1, os.ErrDeadlineExceeded

	case p.pkts <- packet{
		addr: addr,
		buf:  bytes.Clone(buf),
	}:
		return len(buf), nil
	}
}

func (p *PacketPipe) LocalAddr() net.Addr {
	return p.addr
}

func (p *PacketPipe) SetReadDeadline(t time.Time) error {
	p.readDeadline = time.After(time.Until(t))
	return nil
}

func (p *PacketPipe) SetWriteDeadline(t time.Time) error {
	p.writeDeadline = time.After(time.Until(t))
	return nil
}

func (p *PacketPipe) Close() error {
	return nil
}
