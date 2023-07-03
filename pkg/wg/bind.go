// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package wg

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"sync"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	wgconn "golang.zx2c4.com/wireguard/conn"

	"github.com/stv0g/cunicu/pkg/log"
	netx "github.com/stv0g/cunicu/pkg/net"
)

var ErrNoConn = errors.New("no connection for endpoint")

type BindConn interface {
	Receive(buf []byte) (int, wgconn.Endpoint, error)
	Send(buf []byte, ep wgconn.Endpoint) (int, error)

	ListenPort() (uint16, bool)
	SetMark(mark uint32) error

	BindClose() error
}

// BindKernelConn is a BindConn which is consumed by a Kernel WireGuard interface
type BindKernelConn interface {
	BindConn

	WriteKernel([]byte) (int, error)
}

type BindHandler interface {
	OnBindOpen(b *Bind, port uint16)
}

// Compile-time assertion
var _ (wgconn.Bind) = (*Bind)(nil)

type Bind struct {
	Conns []BindConn

	onOpen   []BindHandler
	onPacket []netx.PacketHandler

	endpoints sync.Map

	logger *log.Logger
}

func NewBind(logger *log.Logger) *Bind {
	return &Bind{
		logger: logger.Named("bind"),
	}
}

// Open puts the Bind into a listening state on a given port and reports the actual
// port that it bound to. Passing zero results in a random selection.
// fns is the set of functions that will be called to receive packets.
func (b *Bind) Open(port uint16) ([]wgconn.ReceiveFunc, uint16, error) { //nolint:gocognit
	b.Conns = nil

	for _, h := range b.onOpen {
		h.OnBindOpen(b, port)
	}

	// Add a fallback conn in case no other connections have been registered
	if len(b.Conns) == 0 {
		if err := b.addFallbackConnection(port); err != nil {
			return nil, 0, err
		}
	}

	rcvFns := []wgconn.ReceiveFunc{}
	for _, conn := range b.Conns {
		conn := conn

		rcvFn := func(packets [][]byte, sizes []int, eps []wgconn.Endpoint) (int, error) {
			if len(packets) != 1 {
				panic("batch size not 1?")
			}

			buf := packets[0]

			n, cep, err := conn.Receive(buf)
			if err != nil {
				if errors.Is(err, io.EOF) || errors.Is(err, ice.ErrClosed) || errors.Is(err, net.ErrClosed) {
					err = net.ErrClosed
				} else {
					b.logger.Error("Failed to receive packet", zap.Error(err))
				}

				b.logger.Debug("Connection closed", zap.Error(err))

				return -1, err
			}

			ep := cep.(*BindEndpoint) //nolint:forcetypeassert

			if n > 0 {
				b.logger.Debug("Received packet from bind",
					zap.Int("len", n),
					zap.String("ep", ep.DstToString()),
					zap.Binary("data", buf[:n]))
			}

			sizes[0] = n
			eps[0] = ep

			// Update endpoint map
			if ep.Conn == nil {
				ep.Conn = conn
			}

			// Call handlers
			for _, h := range b.onPacket {
				if abort, err := h.OnPacketRead(buf[:n], ep.DstUDPAddr()); err != nil {
					return -1, fmt.Errorf("failed to call handler: %w", err)
				} else if abort {
					return 0, nil
				}
			}

			return 1, err
		}

		rcvFns = append(rcvFns, rcvFn)
	}

	b.logger.Debug("Opened bind",
		zap.Uint16("port", port),
		zap.Int("#conns", len(rcvFns)))

	return rcvFns, port, nil
}

// Close closes the Bind listener.
// All fns returned by Open must return net.ErrClosed after a call to Close.
func (b *Bind) Close() error {
	for _, c := range b.Conns {
		if err := c.BindClose(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	b.logger.Debug("Closed bind")

	return nil
}

// SetMark sets the mark for each packet sent through this Bind.
// This mark is passed to the kernel as the socket option SO_MARK.
func (b *Bind) SetMark(mark uint32) error {
	b.logger.Debug("Set mark", zap.Uint32("mark", mark))

	for _, c := range b.Conns {
		if err := c.SetMark(mark); err != nil {
			return err
		}
	}

	return nil
}

// Send writes a packet b to address ep.
func (b *Bind) Send(packets [][]byte, cep wgconn.Endpoint) error {
	if len(packets) != 1 {
		panic("batch size not 1?")
	}

	buf := packets[0]
	ep := cep.(*BindEndpoint) //nolint:forcetypeassert

	if ep.Conn == nil {
		return fmt.Errorf("%w: %s", ErrNoConn, ep.DstToString())
	}

	b.logger.Debug("Send packets",
		zap.Int("cnt", len(packets)),
		zap.Any("ep", ep.DstToString()),
		zap.Binary("data", buf))

	_, err := ep.Conn.Send(buf, ep)
	return err
}

// Endpoint returns an Endpoint containing ap.
func (b *Bind) Endpoint(ap netip.AddrPort) *BindEndpoint {
	ep, _ := b.endpoints.LoadOrStore(ap, &BindEndpoint{
		AddrPort: ap,
	})

	return ep.(*BindEndpoint) //nolint:forcetypeassert
}

// ParseEndpoint creates a new endpoint from a string.
// Implements wgconn.Bind
func (b *Bind) ParseEndpoint(s string) (ep wgconn.Endpoint, err error) {
	b.logger.Debug("Parse endpoint", zap.String("ep", s))

	e, err := netip.ParseAddrPort(s)
	return b.Endpoint(e), err
}

// BatchSize is the number of buffers expected to be passed to
// the ReceiveFuncs, and the maximum expected to be passed to SendBatch.
// Implements wgconn.Bind
func (b *Bind) BatchSize() int {
	return 1
}

func (b *Bind) AddConn(conn net.PacketConn) {
	bindConn := newBindPacketConn(b, conn)
	b.Conns = append(b.Conns, bindConn)
}

func (b *Bind) AddOpenHandler(h BindHandler) {
	if !slices.Contains(b.onOpen, h) {
		b.onOpen = append(b.onOpen, h)
	}
}

func (b *Bind) RemoveOpenHandler(h BindHandler) {
	if idx := slices.Index(b.onOpen, h); idx > -1 {
		b.onOpen = slices.Delete(b.onOpen, idx, idx+1)
	}
}

func (b *Bind) AddPacketHandler(h netx.PacketHandler) {
	if !slices.Contains(b.onPacket, h) {
		b.onPacket = append(b.onPacket, h)
	}
}

func (b *Bind) RemovePacketHandler(h netx.PacketHandler) {
	if idx := slices.Index(b.onPacket, h); idx > -1 {
		b.onPacket = slices.Delete(b.onPacket, idx, idx+1)
	}
}

func (b *Bind) addFallbackConnection(port uint16) error {
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: int(port),
	})
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	b.AddConn(udpConn)

	return nil
}
