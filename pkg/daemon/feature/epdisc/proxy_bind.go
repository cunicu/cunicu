// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	wgconn "golang.zx2c4.com/wireguard/conn"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/wg"
)

// Compile-time assertions
var (
	_ Proxy       = (*BindProxy)(nil)
	_ wg.BindConn = (*BindProxy)(nil)
)

type BindProxy struct {
	ep *wg.BindEndpoint

	iceConn *ice.Conn

	logger *log.Logger
}

func NewBindProxy(bind *wg.Bind, cp *ice.CandidatePair, conn *ice.Conn, logger *log.Logger) (*BindProxy, *net.UDPAddr, error) {
	p := &BindProxy{
		iceConn: conn,

		logger: logger.Named("proxy").With(zap.String("type", "bind")),
	}

	epAddr := &net.UDPAddr{
		IP:   net.ParseIP(cp.Remote.Address()),
		Port: cp.Remote.Port(),
	}

	if v4 := epAddr.IP.To4(); v4 != nil {
		epAddr.IP = v4
	}

	p.ep = bind.Endpoint(epAddr.AddrPort())
	p.ep.Conn = p

	return p, epAddr, nil
}

func (p *BindProxy) Close() error {
	return nil
}

func (p *BindProxy) BindClose() error {
	// Unblock Read() in Receive()
	return p.iceConn.SetReadDeadline(time.Now())
}

// The following methods implement wg.BindConn

func (p *BindProxy) Receive(buf []byte) (int, wgconn.Endpoint, error) {
	n, err := p.iceConn.Read(buf)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrDeadlineExceeded):
			// We use the deadline exceeded just to manually unblock Receive()
			// instead of really closing the connection. So lets fake a closed
			// connection here.
			err = net.ErrClosed

		case errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF):
			p.logger.DebugV(10, "Connection closed. Returning from receive()")

		default:
			p.logger.Error("Failed to read", zap.Error(err))
		}

		return -1, nil, err
	}

	return n, p.ep, nil
}

func (p *BindProxy) Send(buf []byte, ep wgconn.Endpoint) (int, error) {
	if p.ep != ep {
		return -1, fmt.Errorf("%w: %s != %s", errMismatchingEndpoints, p.ep.DstToString(), ep.DstToString())
	}

	return p.iceConn.Write(buf)
}

func (p *BindProxy) ListenPort() (uint16, bool) {
	return 0, false
}

func (p *BindProxy) SetMark(_ uint32) error {
	return errNotSupported
}
