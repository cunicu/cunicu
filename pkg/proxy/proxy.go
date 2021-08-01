package proxy

import (
	"errors"
	"io"
	"net"
)

type ProxyType int

type UpdateEndpointCb func(addr *net.UDPAddr) error

const (
	ProxyTypeInvalid ProxyType = iota
	ProxyTypeAuto
	ProxyTypeUser
	ProxyTypeNFTables
	ProxyTypeEBPF

	StunMagicCookie uint32 = 0x2112A442
)

type Proxy interface {
	io.Closer

	UpdateEndpoint(addr *net.UDPAddr) error
}

type BaseProxy struct {
	ListenPort int
	Ident      string
}

func ProxyTypeFromString(typ string) ProxyType {
	switch typ {
	case "auto":
		return ProxyTypeAuto
	case "user":
		return ProxyTypeUser
	case "nftables":
		return ProxyTypeNFTables
	case "ebpf":
		return ProxyTypeEBPF
	default:
		return ProxyTypeInvalid
	}
}

func (pt ProxyType) String() string {
	switch pt {
	case ProxyTypeAuto:
		return "auto"
	case ProxyTypeUser:
		return "user"
	case ProxyTypeNFTables:
		return "nftables"
	case ProxyTypeEBPF:
		return "ebpf"
	}

	return "invalid"
}

func AutoProxy() ProxyType {
	if CheckEBPFSupport() {
		return ProxyTypeEBPF
	} else if CheckNFTablesSupport() {
		return ProxyTypeNFTables
	} else {
		return ProxyTypeUser
	}
}

func NewProxy(pt ProxyType, ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (Proxy, error) {
	switch pt {
	case ProxyTypeUser:
		return NewUserProxy(ident, listenPort, cb, conn)
	case ProxyTypeNFTables:
		return NewNFTablesProxy(ident, listenPort, cb, conn)
	case ProxyTypeEBPF:
		return NewEBPFProxy(ident, listenPort, cb, conn)
	}

	return nil, errors.New("unknown proxy type")
}

func Type(p Proxy) ProxyType {
	switch p.(type) {
	case *NFTablesProxy:
		return ProxyTypeNFTables
	case *UserProxy:
		return ProxyTypeUser
	case *EBPFProxy:
		return ProxyTypeEBPF
	default:
		return ProxyTypeInvalid
	}
}
