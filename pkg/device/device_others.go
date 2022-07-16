//go:build !linux

package device

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/internal/config"
	"riasc.eu/wice/internal/errors"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

func NewKernelDevice(name string) (KernelDevice, error) {
	return nil, errors.ErrNotSupported
}
