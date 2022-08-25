package proxy

import (
	"io"
	"net"

	"github.com/pion/ice/v2"
	"riasc.eu/wice/pkg/pb"
)

type Proxy interface {
	io.Closer

	Update(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error)

	Type() pb.ProxyType
}
