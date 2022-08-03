//go:build !(linux || freebsd || darwin)

package device

import (
	"riasc.eu/wice/pkg/errors"
)

func FindDevice(name string) (KernelDevice, error) {
	return nil, errors.ErrNotSupported
}

func NewKernelDevice(name string) (KernelDevice, error) {
	return nil, errors.ErrNotSupported
}
