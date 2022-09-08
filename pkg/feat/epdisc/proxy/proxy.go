// Package proxy provides tooling for transparently proxying STUN/TURN trafic between ICE agents and kernel/userspace WireGuard interfaces
package proxy

import (
	"io"
	"net"

	"github.com/pion/ice/v2"

	protoepdisc "github.com/stv0g/cunicu/pkg/proto/feat/epdisc"
)

type Proxy interface {
	io.Closer

	UpdateCandidatePair(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error)

	Type() protoepdisc.ProxyType
}
