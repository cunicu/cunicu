package proxy

import (
	"io"
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
)

const (
	maxSegmentSize = (1 << 16) - 1
)

type UserProxy struct {
	BaseProxy
	conn *net.UDPConn
}

func NewUserProxy(ident string, listenPort int, cb UpdateEndpointCb, conn net.Conn) (Proxy, error) {
	var err error

	logger := zap.L().Named("proxy.user")

	p := &UserProxy{
		BaseProxy: BaseProxy{
			Ident:  ident,
			logger: logger,
		},
	}

	// Userspace proxying
	rAddr := net.UDPAddr{
		IP:   nil, // localhost
		Port: listenPort,
	}
	lAddr := net.UDPAddr{
		IP:   net.IPv6loopback,
		Port: 0, // choose automatically
	}

	p.conn, err = net.DialUDP("udp", &lAddr, &rAddr)
	if err != nil {
		return nil, err
	}

	// Update Wireguard peer endpoint
	addr := p.conn.LocalAddr().(*net.UDPAddr)
	if err := cb(addr); err != nil {
		return nil, err
	}

	ingressBuf := make([]byte, maxSegmentSize)
	egressBuf := make([]byte, maxSegmentSize)

	// Bi-directional copy between ICE and loopback UDP sockets until proxy.conn is closed
	go io.CopyBuffer(conn, p.conn, ingressBuf)
	go io.CopyBuffer(p.conn, conn, egressBuf)

	p.logger.Info("Setup user-space proxy")

	return p, nil
}

func (p *UserProxy) Type() ProxyType {
	return TypeUser
}

func (p *UserProxy) Setup(agentConfig *ice.AgentConfig, listenPort int) error {
	return nil
}

func (p *UserProxy) Close() error {
	return p.conn.Close()
}
