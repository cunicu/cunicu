//go:build !(android || linux || freebsd || openbsd)

package wg

import (
	"net"
)

const fwmarkIoctl = 0

func SetMark(conn net.PacketConn, mark uint32) error {
	return errNotSupported
}
