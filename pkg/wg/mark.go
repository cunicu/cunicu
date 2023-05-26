// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build android || linux || openbsd || freebsd

package wg

import (
	"net"

	"golang.org/x/sys/unix"
)

func SetMark(conn net.PacketConn, mark uint32) error {
	var operr error

	if fwmarkIoctl == 0 {
		return errNotSupported
	}

	udpConn, ok := conn.(*net.UDPConn)
	if !ok {
		return errNotSupported
	}

	rawConn, err := udpConn.SyscallConn()
	if err != nil {
		return err
	}

	if err = rawConn.Control(func(fd uintptr) {
		operr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, fwmarkIoctl, int(mark))
	}); err == nil {
		err = operr
	}
	if err != nil {
		return err
	}

	return nil
}
