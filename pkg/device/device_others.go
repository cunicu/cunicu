//go:build !(linux || freebsd || darwin)

package device

import (
	"riasc.eu/wice/pkg/errors"
)

func FindKernelDevice(name string) (Device, error) {
	return nil, errors.ErrNotSupported
}

func NewKernelDevice(name string) (Device, error) {
	return nil, errors.ErrNotSupported
}
