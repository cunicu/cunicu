package device

import (
	"errors"
	"fmt"
	"net"

	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/wg"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/conn"
	wgdevice "golang.zx2c4.com/wireguard/device"
)

type KernelDevice struct {
	link.Link

	ListenPort int
	bind       *wg.Bind

	logger *zap.Logger
}

func NewKernelDevice(name string) (*KernelDevice, error) {
	logger := zap.L().Named("dev").With(
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
	logger := zap.L().Named("dev").With(
		zap.String("dev", name),
		zap.String("type", "kernel"),
	)

	lnk, err := link.FindLink(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find WireGuard link: %w", err)
	}

	// TODO: Is this portable?
	if lnk.Type() != link.TypeWireGuard {
		return nil, fmt.Errorf("link '%s' is not a WireGuard link", lnk.Name())
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
		d.logger.Warn("Skip bind update as we no listen port yet")
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

func (d *KernelDevice) doReceive(rcvFn conn.ReceiveFunc) {
	d.logger.Debug("Receive worker started")

	buf := make([]byte, wgdevice.MaxMessageSize)

	for {
		n, cep, err := rcvFn(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}

			d.logger.Error("Failed to receive from bind", zap.Error(err))
			continue
		} else if n == 0 {
			continue
		}

		ep := cep.(*wg.BindEndpoint) //nolint:forcetypeassert
		kc, ok := ep.Conn.(wg.BindKernelConn)
		if !ok {
			d.logger.Error("No kernel connection found", zap.String("ep", ep.DstToString()))
			continue
		}

		if _, err := kc.WriteKernel(buf[:n]); err != nil {
			d.logger.Error("Failed to write to kernel", zap.Error(err))
		}
	}

	d.logger.Debug("Receive worker stopped")
}
