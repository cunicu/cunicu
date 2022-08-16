package proxy

import (
	"io"
	"net"

	"github.com/pion/ice/v2"
)

type Proxy interface {
	io.Closer

	Update(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error)
}
