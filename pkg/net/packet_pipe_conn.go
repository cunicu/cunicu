// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"net"
	"time"
)

type PacketPipeConn struct {
	rx *PacketPipe
	tx *PacketPipe
}

func NewPacketPipeConn(l1 net.Addr, l2 net.Addr, depth int) (*PacketPipeConn, *PacketPipeConn) {
	p1 := NewPacketPipe(l1, depth)
	p2 := NewPacketPipe(l2, depth)

	pc1 := &PacketPipeConn{
		rx: p1,
		tx: p2,
	}

	pc2 := &PacketPipeConn{
		rx: p2,
		tx: p1,
	}

	return pc1, pc2
}

func (c *PacketPipeConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return c.rx.ReadFrom(p)
}

func (c *PacketPipeConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return c.tx.WriteFrom(p, c.rx.LocalAddr())
}

func (c *PacketPipeConn) Close() error {
	if err := c.rx.Close(); err != nil {
		return err
	}

	return c.tx.Close()
}

func (c *PacketPipeConn) LocalAddr() net.Addr {
	return c.rx.LocalAddr()
}

func (c *PacketPipeConn) SetDeadline(t time.Time) error {
	if err := c.rx.SetReadDeadline(t); err != nil {
		return err
	}

	return c.tx.SetWriteDeadline(t)
}

func (c *PacketPipeConn) SetReadDeadline(t time.Time) error {
	return c.rx.SetReadDeadline(t)
}

func (c *PacketPipeConn) SetWriteDeadline(t time.Time) error {
	return c.tx.SetWriteDeadline(t)
}
