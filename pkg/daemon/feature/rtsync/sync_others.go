// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package rtsync

import (
	"github.com/stv0g/cunicu/pkg/daemon"
)

func (i *Interface) removeKernel(_ *daemon.Peer) error {
	return errNotSupported
}

func (i *Interface) syncKernel() error {
	return errNotSupported
}

func (i *Interface) watchKernel() error {
	return errNotSupported
}
