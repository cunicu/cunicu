// Package proxy provides tooling for transparently forwarding STUN/TURN traffic
// between ICE agents and kernel/userspace WireGuard interfaces
package proxy

import (
	"io"
	"net"

	"github.com/pion/ice/v2"

	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

type Proxy interface {
	io.Closer

	UpdateCandidatePair(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error)

	Type() epdiscproto.ProxyType
}
