// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"net"
	"time"
)

type ReceivePacketConn interface {
	ReadFrom(p []byte) (n int, addr net.Addr, err error)
	Close() error
	LocalAddr() net.Addr
	SetReadDeadline(t time.Time) error
}

type SendPacketConn interface {
	WriteTo(p []byte, addr net.Addr) (n int, err error)
	Close() error
	SetWriteDeadline(t time.Time) error
}

type SplitConn struct {
	recv ReceivePacketConn
	send SendPacketConn
}

// Compile-time assertions
var _ net.PacketConn = (*SplitConn)(nil)

func NewSplitConn(recv ReceivePacketConn, send SendPacketConn) net.PacketConn {
	return &SplitConn{
		recv: recv,
		send: send,
	}
}

func (c *SplitConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	return c.recv.ReadFrom(p)
}

func (c *SplitConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	return c.send.WriteTo(p, addr)
}

func (c *SplitConn) Close() error {
	if err := c.recv.Close(); err != nil {
		return err
	}

	return c.send.Close()
}

func (c *SplitConn) LocalAddr() net.Addr {
	return c.recv.LocalAddr()
}

func (c *SplitConn) SetDeadline(t time.Time) error {
	if err := c.recv.SetReadDeadline(t); err != nil {
		return err
	}

	return c.send.SetWriteDeadline(t)
}

func (c *SplitConn) SetReadDeadline(t time.Time) error {
	return c.recv.SetReadDeadline(t)
}

func (c *SplitConn) SetWriteDeadline(t time.Time) error {
	return c.send.SetWriteDeadline(t)
}
