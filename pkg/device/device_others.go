//go:build !(linux || freebsd || darwin)

package device

import (
	"github.com/stv0g/cunicu/pkg/errors"
)

func FindKernelDevice(name string) (Device, error) {
	return nil, errors.ErrNotSupported
}

func NewKernelDevice(name string) (Device, error) {
	return nil, errors.ErrNotSupported
}
