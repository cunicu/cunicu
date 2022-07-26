//go:build !linux

package device

import (
	"golang.zx2c4.com/wireguard/wgctrl"
	"riasc.eu/wice/pkg/config"
	"riasc.eu/wice/pkg/errors"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

func NewKernelDevice(name string) (KernelDevice, error) {
	return nil, errors.ErrNotSupported
}
