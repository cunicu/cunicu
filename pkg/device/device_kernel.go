// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package device

import (
	"errors"
	"fmt"
	"net"

	"go.uber.org/zap"
	wgconn "golang.zx2c4.com/wireguard/conn"
	wgdevice "golang.zx2c4.com/wireguard/device"

	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/log"
	"github.com/stv0g/cunicu/pkg/wg"
)

var errNotWireGuardLink = errors.New("link is not a WireGuard link")

type KernelDevice struct {
	link.Link

	ListenPort int
	bind       *wg.Bind

	logger *log.Logger
}

func NewKernelDevice(name string) (*KernelDevice, error) {
	logger := log.Global.Named("dev").With(
		zap.String("dev", name),
		zap.String("type", "kernel"),
	)

	lnk, err := link.CreateWireGuardLink(name)
	if err != nil {
		return nil, fmt.Errorf("failed to create WireGuard link: %w", err)
	}

	return &KernelDevice{
		Link:   lnk,
		bind:   wg.NewBind(logger),
		logger: logger,
	}, nil
}

func FindKernelDevice(name string) (*KernelDevice, error) {
	logger := log.Global.Named("dev").With(
		zap.String("dev", name),
		zap.String("type", "kernel"),
	)

	lnk, err := link.FindLink(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find WireGuard link: %w", err)
	}

	// TODO: Is this portable?
	if lnk.Type() != link.TypeWireGuard {
		return nil, fmt.Errorf("%w: %s", errNotWireGuardLink, lnk.Name())
	}

	return &KernelDevice{
		Link:   lnk,
		bind:   wg.NewBind(logger),
		logger: logger,
	}, nil
}

func (d *KernelDevice) Bind() *wg.Bind {
	return d.bind
}

func (d *KernelDevice) BindUpdate() error {
	if d.ListenPort == 0 {
		d.logger.Debug("Skip bind update as we no listen port yet")
		return nil
	}

	if err := d.bind.Close(); err != nil {
		return fmt.Errorf("failed to close bind: %w", err)
	}

	rcvFns, _, err := d.bind.Open(0)
	if err != nil {
		return fmt.Errorf("failed to open bind: %w", err)
	}

	for _, rcvFn := range rcvFns {
		go d.doReceive(rcvFn)
	}

	return nil
}

func (d *KernelDevice) doReceive(rcvFn wgconn.ReceiveFunc) {
	d.logger.Debug("Receive worker started")

	batchSize := 1
	packets := make([][]byte, batchSize)
	sizes := make([]int, batchSize)
	eps := make([]wgconn.Endpoint, batchSize)

	packets[0] = make([]byte, wgdevice.MaxMessageSize)

	for {
		n, err := rcvFn(packets, sizes, eps)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			d.logger.Error("Failed to receive from bind", zap.Error(err))
			continue
		} else if n == 0 || sizes[0] == 0 {
			continue
		}

		ep := eps[0].(*wg.BindEndpoint) //nolint:forcetypeassert
		kc, ok := ep.Conn.(wg.BindKernelConn)
		if !ok {
			d.logger.Error("No kernel connection found", zap.String("ep", ep.DstToString()))
			continue
		}

		if _, err := kc.WriteKernel(packets[0][:sizes[0]]); err != nil {
			d.logger.Error("Failed to write to kernel", zap.Error(err))
		}
	}

	d.logger.Debug("Receive worker stopped")
}
