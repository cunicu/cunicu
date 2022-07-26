//go:build !linux

package proxy

import (
	"riasc.eu/wice/pkg/errors"
)

import (
	"net"
)

func createFilteredSTUNConnection(listenPort int) (net.PacketConn, error) {
	return nil, errors.ErrNotSupported
}
