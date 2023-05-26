// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !(android || linux || freebsd || openbsd)

package wg

import (
	"net"
)

func SetMark(_ net.PacketConn, _ uint32) error {
	return errNotSupported
}
