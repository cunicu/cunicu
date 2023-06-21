// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/stv0g/cunicu/pkg/log"
	"go.uber.org/zap"
	wgconn "golang.zx2c4.com/wireguard/conn"
)

var _ BindConn = (*BindPacketConn)(nil)

// BindPacketConn is a PacketConn
type BindPacketConn struct {
	net.PacketConn

	bind   *Bind
	logger *log.Logger
}

func NewBindPacketConn(bind *Bind, conn net.PacketConn, logger *log.Logger) *BindPacketConn {
	return &BindPacketConn{
		PacketConn: conn,
		bind:       bind,
		logger:     logger,
	}
}

func (c *BindPacketConn) Receive(buf []byte) (int, wgconn.Endpoint, error) {
	// Reset read deadline
	if err := c.PacketConn.SetReadDeadline(time.Time{}); err != nil {
		return -1, nil, fmt.Errorf("failed to reset read deadline: %w", err)
	}

	n, rAddr, err := c.PacketConn.ReadFrom(buf)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrDeadlineExceeded):
			// We use the deadline exceeded just to manually unblock Receive()
			// instead of really closing the connection. So lets fake a closed
			// connection here.
			err = net.ErrClosed

		case errors.Is(err, net.ErrClosed):
			c.logger.Debug("Connection closed. Returning from receive()")

		default:
			c.logger.Error("Failed to read", zap.Error(err))
		}

		return -1, nil, err
	}

	rUDPAddr, ok := rAddr.(*net.UDPAddr)
	if !ok {
		panic("failed to cast")
	}

	if v4 := rUDPAddr.IP.To4(); v4 != nil {
		rUDPAddr.IP = v4
	}

	udpAddrPort := rUDPAddr.AddrPort()
	ep := c.bind.Endpoint(udpAddrPort)

	if ep.Conn == nil {
		ep.Conn = c
	}

	return n, ep, nil
}

func (c *BindPacketConn) Send(buf []byte, cep wgconn.Endpoint) (int, error) {
	ep := cep.(*BindEndpoint) //nolint:forcetypeassert

	return c.PacketConn.WriteTo(buf, ep.DstUDPAddr())
}

func (c *BindPacketConn) SetMark(mark uint32) error {
	return SetMark(c.PacketConn, mark)
}

func (c *BindPacketConn) ListenPort() (uint16, bool) {
	if addr, ok := c.PacketConn.LocalAddr().(*net.UDPAddr); ok {
		return uint16(addr.Port), true
	}

	return 0, false
}

func (c *BindPacketConn) BindClose() error {
	// We do not want to close the underlying connections here
	// as this would disrupt Pion's UDPMux's which would
	// need to be recreated.
	// Instead which just unblock the currently blocking receive
	// functions by calling conn.SetReadDeadline()

	return c.PacketConn.SetReadDeadline(time.Now())
}
