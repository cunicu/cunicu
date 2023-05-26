// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package device implements OS abstractions for managing WireGuard links
package device

import (
	"github.com/stv0g/cunicu/pkg/link"
	"github.com/stv0g/cunicu/pkg/wg"
)

type Device interface {
	link.Link

	Bind() *wg.Bind
	BindUpdate() error
}

func NewDevice(name string, user bool) (kernelDev Device, err error) {
	if user {
		kernelDev, err = NewUserDevice(name)
	} else {
		kernelDev, err = NewKernelDevice(name)
	}
	if err != nil {
		return nil, err
	}

	return kernelDev, nil
}
