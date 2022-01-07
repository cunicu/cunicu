//go:build !linux

package proxy

import (
	"errors"
	"net"

	"github.com/pion/ice/v2"
)

func SetupEBPFProxy(cfg *ice.AgentConfig, port int) error {
	return errors.New("the eBPF proxy mode is unsupported on this system")
}

func NewEBPFProxy(ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (Proxy, error) {
	return nil, errors.New("the eBPF proxy mode is unsupported on this system")
}

func NewNFTablesProxy(ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (Proxy, error) {
	return nil, errors.New("the eBPF proxy mode is unsupported on this system")
}
