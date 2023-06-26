// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
	wgconn "golang.zx2c4.com/wireguard/conn"

	"github.com/stv0g/cunicu/pkg/log"
)

var _ BindConn = (*bindPacketConn)(nil)

// bindPacketConn is a PacketConn
type bindPacketConn struct {
	net.PacketConn

	bind   *Bind
	logger *log.Logger
}

func newBindPacketConn(bind *Bind, conn net.PacketConn) *bindPacketConn {
	return &bindPacketConn{
		PacketConn: conn,
		bind:       bind,
		logger:     bind.logger.Named("conn"),
	}
}

func (c *bindPacketConn) Receive(buf []byte) (int, wgconn.Endpoint, error) {
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
			c.logger.DebugV(10, "Connection closed. Returning from receive()")

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

func (c *bindPacketConn) Send(buf []byte, cep wgconn.Endpoint) (int, error) {
	ep := cep.(*BindEndpoint) //nolint:forcetypeassert

	return c.PacketConn.WriteTo(buf, ep.DstUDPAddr())
}

func (c *bindPacketConn) SetMark(mark uint32) error {
	return SetMark(c.PacketConn, mark)
}

func (c *bindPacketConn) ListenPort() (uint16, bool) {
	if addr, ok := c.PacketConn.LocalAddr().(*net.UDPAddr); ok {
		return uint16(addr.Port), true
	}

	return 0, false
}

func (c *bindPacketConn) BindClose() error {
	// We do not want to close the underlying connections here
	// as this would disrupt Pion's UDPMux's which would
	// need to be recreated.
	// Instead which just unblock the currently blocking receive
	// functions by calling conn.SetReadDeadline()

	return c.PacketConn.SetReadDeadline(time.Now())
}
