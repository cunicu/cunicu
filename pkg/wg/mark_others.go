//go:build !(android || linux || freebsd || openbsd)

package wg

import (
	"net"
)

func SetMark(_ net.PacketConn, _ uint32) error {
	return errNotSupported
}
