//go:build android || linux || openbsd || freebsd

package wg

import (
	"net"

	"golang.org/x/sys/unix"
)

func SetMark(conn *net.UDPConn, mark uint32) error {
	var operr error

	if fwmarkIoctl == 0 {
		return errNotSupported
	}

	rawConn, err := conn.SyscallConn()
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
