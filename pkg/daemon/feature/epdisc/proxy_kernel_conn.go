// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package epdisc

import (
	"errors"
	"fmt"
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
	wgdevice "golang.zx2c4.com/wireguard/device"

	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/wg"
)

// Compile-time assertions
var (
	_ Proxy             = (*KernelConnProxy)(nil)
	_ wg.BindKernelConn = (*KernelConnProxy)(nil)
)

var errNotSupported = errors.New("not supported")

type KernelConnProxy struct {
	*BindProxy

	kernelConn *net.UDPConn

	logger *log.Logger
}

func NewKernelConnProxy(bind *wg.Bind, cp *ice.CandidatePair, conn *ice.Conn, listenPort int, logger *log.Logger) (*KernelConnProxy, *net.UDPAddr, error) {
	bp, _, err := NewBindProxy(bind, cp, conn, logger)
	if err != nil {
		return nil, nil, err
	}

	p := &KernelConnProxy{
		BindProxy: bp,

		logger: logger.Named("proxy").With(zap.String("type", "kernel")),
	}

	p.ep.Conn = p

	lAddr := &net.UDPAddr{
		IP: net.IPv6loopback,
	}
	rAddr := &net.UDPAddr{
		IP:   net.IPv6loopback,
		Port: listenPort,
	}

	p.kernelConn, err = net.DialUDP("udp", lAddr, rAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := p.forwardFromKernel(); err != nil {
			p.logger.Error("Failed to forward from kernel", zap.Error(err))
		}
	}()

	epAddr := p.kernelConn.LocalAddr().(*net.UDPAddr) //nolint:forcetypeassert

	return p, epAddr, nil
}

// Close releases all resources of the proxy
func (p *KernelConnProxy) Close() error {
	p.logger.Debug("Closing kernel connection")
	if err := p.kernelConn.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return p.BindProxy.Close()
}

func (p *KernelConnProxy) forwardFromKernel() error {
	buf := make([]byte, wgdevice.MaxMessageSize)

	p.logger.Debug("Start forwarding from kernel to ICE connection")

	for {
		n, err := p.kernelConn.Read(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			return fmt.Errorf("failed to read from kernel: %w", err)
		}

		p.logger.Debug("Received from kernel", zap.Int("n", n), zap.Binary("buf", buf[:n]))

		if _, err := p.iceConn.Write(buf[:n]); err != nil {
			return fmt.Errorf("failed to write to ICE conn: %w", err)
		}
	}

	p.logger.Debug("Stopped forwarding from kernel to ICE connection")

	return nil
}

// The following functions implement wg.BindKernelConn

func (p *KernelConnProxy) WriteKernel(b []byte) (int, error) {
	n, err := p.kernelConn.Write(b)

	p.logger.Debug("Send to kernel", zap.Int("n", n), zap.Binary("buf", b[:n]))

	return n, err
}
