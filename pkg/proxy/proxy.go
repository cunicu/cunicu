package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"runtime"

	"go.uber.org/zap"
)

type ProxyType int

type UpdateEndpointCb func(addr *net.UDPAddr) error

const (
	TypeInvalid ProxyType = iota
	TypeAuto
	TypeUser
	TypeNFTables
	TypeEBPF

	StunMagicCookie uint32 = 0x2112A442
)

type Proxy interface {
	io.Closer

	Type() ProxyType
}

type BaseProxy struct {
	ListenPort int
	Ident      string
	logger     *zap.Logger
}

func CheckNFTablesSupport() bool {
	return runtime.GOOS == "linux"
}

func CheckEBPFSupport() bool {
	return runtime.GOOS == "linux"
}

func ProxyTypeFromString(typ string) (ProxyType, error) {
	switch typ {
	case "auto":
		return TypeAuto, nil
	case "user":
		return TypeUser, nil
	case "nftables":
		return TypeNFTables, nil
	case "ebpf":
		return TypeEBPF, nil
	default:
		return -1, fmt.Errorf("invalid proxy type: %s", typ)
	}
}

func (pt ProxyType) String() string {
	switch pt {
	case TypeAuto:
		return "auto"
	case TypeUser:
		return "user"
	case TypeNFTables:
		return "nftables"
	case TypeEBPF:
		return "ebpf"
	}

	return "invalid"
}

func AutoProxy() ProxyType {
	if CheckEBPFSupport() {
		return TypeEBPF
	} else if CheckNFTablesSupport() {
		return TypeNFTables
	} else {
		return TypeUser
	}
}

func NewProxy(pt ProxyType, ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (Proxy, error) {
	switch pt {
	case TypeUser:
		return NewUserProxy(ident, listenPort, cb, conn)
	case TypeNFTables:
		return NewNFTablesProxy(ident, listenPort, cb, conn)
	case TypeEBPF:
		return NewEBPFProxy(ident, listenPort, cb, conn)
	}

	return nil, errors.New("unknown proxy type")
}
