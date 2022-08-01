package proxy

import (
	"fmt"
	"io"
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"
)

const (
	StunMagicCookie uint32 = 0x2112A442

	maxSegmentSize = (1 << 16) - 1
)

type Proxy struct {
	listenPort int

	nat *NAT

	iceConn  *ice.Conn
	userConn *net.UDPConn

	logger *zap.Logger
}

func NewProxy(nat *NAT, listenPort int) (*Proxy, error) {
	p := &Proxy{
		nat:        nat,
		listenPort: listenPort,
		logger:     zap.L().Named("proxy"),
	}

	return p, nil
}

func (p *Proxy) Close() error {
	if p.userConn != nil {
		if err := p.userConn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (p *Proxy) Update(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error) {
	// By default we proxy through the userspace
	var ep *net.UDPAddr

	if cp.Local.Type() == ice.CandidateTypeHost || cp.Local.Type() == ice.CandidateTypeServerReflexive {
		ep = &net.UDPAddr{
			IP:   net.ParseIP(cp.Remote.Address()),
			Port: cp.Remote.Port(),
		}

		// Update SNAT set for UDPMuxSrflx
		if err := p.nat.MasqueradeSourcePort(p.listenPort, cp.Local.Port(), ep); err != nil {
			return nil, err
		}
	} else {
		// We cant to anything for prfx and relay candidates.
		// Let them pass through the userspace connection

		// We create the user connection only on demand to avoid opening unused sockets
		if p.userConn == nil {
			if err := p.setupUserConn(conn); err != nil {
				return nil, fmt.Errorf("failed to setup user connection: %w", err)
			}
		}

		// Start copying if the underlying ice.Conn has changed
		if conn != p.iceConn {
			p.iceConn = conn

			// Bi-directional copy between ICE and loopback UDP sockets
			go p.copy(conn, p.userConn)
			go p.copy(p.userConn, conn)
		}

		ep = p.userConn.LocalAddr().(*net.UDPAddr)
	}

	return ep, nil
}

func (p *Proxy) copy(dst io.Writer, src io.Reader) {
	buf := make([]byte, maxSegmentSize)
	for {
		// if _, err := io.Copy(dst, src); err != nil {
		// 	p.logger.Error("Failed copy", zap.Error(err))
		// }

		n, err := src.Read(buf)
		if err != nil {
			p.logger.Error("Failed to read", zap.Error(err))
			continue
		}

		_, err = dst.Write(buf[:n])
		if err != nil {
			p.logger.Error("Failed to write", zap.Error(err))
		}
	}
}

func (p *Proxy) setupUserConn(iceConn *ice.Conn) error {
	var err error

	// User-space proxying
	rAddr := net.UDPAddr{
		IP:   net.IPv6loopback,
		Port: int(p.listenPort),
	}
	lAddr := net.UDPAddr{
		IP:   net.IPv6loopback,
		Port: 0, // choose randomly
	}

	if p.userConn, err = net.DialUDP("udp", &lAddr, &rAddr); err != nil {
		return err
	}

	p.logger.Info("Setup user-space proxy",
		zap.Any("localAddress", p.userConn.LocalAddr()),
		zap.Any("remoteAddress", p.userConn.RemoteAddr()))

	return nil
}
