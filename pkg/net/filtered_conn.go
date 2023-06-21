// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package net

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/log"
)

var ErrFiltered = errors.New("packet has been filtered")

// PacketHandler is a handler interface
type PacketHandler interface {
	OnPacketRead([]byte, net.Addr) (bool, error)
}

// FilteredConn wraps a net.PacketConn
type FilteredConn struct {
	net.PacketConn

	onPacket []PacketHandler
	logger   *log.Logger
}

func NewFilteredConn(c net.PacketConn, logger *log.Logger) *FilteredConn {
	return &FilteredConn{
		PacketConn: c,
		logger:     logger,
	}
}

func (c *FilteredConn) ReadFrom(buf []byte) (int, net.Addr, error) {
out:
	for {
		n, rAddr, err := c.PacketConn.ReadFrom(buf)
		if err != nil {
			return -1, nil, err
		}

		// Call handlers
		for _, h := range c.onPacket {
			if abort, err := h.OnPacketRead(buf[:n], rAddr); err != nil {
				return -1, nil, fmt.Errorf("failed to call handler: %w", err)
			} else if abort {
				continue out
			}
		}

		return n, rAddr, nil
	}
}

func (c *FilteredConn) AddPacketReadHandler(h PacketHandler) {
	if !slices.Contains(c.onPacket, h) {
		c.onPacket = append(c.onPacket, h)
	}
}

func (c *FilteredConn) RemovePacketReadHandler(h PacketHandler) {
	if idx := slices.Index(c.onPacket, h); idx > 0 {
		slices.Delete(c.onPacket, idx, idx+1)
	}
}

func (c *FilteredConn) AddPacketReadHandlerConn(h PacketHandler) net.PacketConn {
	hc := &PacketHandlerConn{
		PacketHandler: h,
		pipe:          NewPacketPipe(c.LocalAddr(), 1024),
	}

	c.AddPacketReadHandler(hc)

	return NewSplitConn(hc.pipe, c.PacketConn)
}

// PacketHandlerConn implements a PacketHandler which forwards
// filtered reads to a pipe connection
type PacketHandlerConn struct {
	PacketHandler

	pipe *PacketPipe
}

func (ph *PacketHandlerConn) OnPacketRead(buf []byte, rAddr net.Addr) (abort bool, err error) {
	if abort, err = ph.PacketHandler.OnPacketRead(buf, rAddr); err != nil {
		return false, err
	} else if abort {
		buf := slices.Clone(buf)
		_, err := ph.pipe.WriteFrom(buf, rAddr)

		return true, err
	}

	return abort, nil
}
