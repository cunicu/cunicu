package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"

	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

const (
	StunMagicCookie uint32 = 0x2112A442

	maxSegmentSize = (1 << 16) - 1
)

type KernelProxy struct {
	listenPort int

	nat     *NAT
	natRule *NATRule

	ep       *net.UDPAddr
	cp       *ice.CandidatePair
	connICE  *ice.Conn
	connUser *net.UDPConn

	logger *zap.Logger
}

func NewKernelProxy(nat *NAT, listenPort int) (*KernelProxy, error) {
	p := &KernelProxy{
		nat:        nat,
		listenPort: listenPort,
		logger:     zap.L().Named("proxy").With(zap.String("type", "kernel")),
	}

	return p, nil
}

func (p *KernelProxy) Close() error {
	if p.connUser != nil {
		p.connUser.SetWriteDeadline(time.Now().Add(1 * time.Second)) // TODO: really required?
		if err := p.connUser.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (p *KernelProxy) UpdateListenPort(listenPort int) error {
	return p.Update(nil, nil, listenPort)
}

func (p *KernelProxy) UpdateCandidatePair(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error) {
	if err := p.Update(cp, conn, -1); err != nil {
		return nil, err
	}

	return p.ep, nil
}

func (p *KernelProxy) Update(newCP *ice.CandidatePair, newConnICE *ice.Conn, newListenPort int) error {
	var err error

	if newListenPort > 0 {
		p.listenPort = newListenPort
	}

	if newCP != nil {
		p.cp = newCP
	}

	if p.cp != nil {
		switch p.Type() {
		case epdiscproto.ProxyType_KERNEL_NAT:
			p.ep = &net.UDPAddr{
				IP:   net.ParseIP(p.cp.Remote.Address()),
				Port: p.cp.Remote.Port(),
			}

			// Delete any old SPAT rule
			if p.natRule != nil {
				if err := p.natRule.Delete(); err != nil {
					return fmt.Errorf("failed to delete rule: %w", err)
				}
			}

			// Setup SNAT redirect (WireGuard listen-port -> STUN port)
			if p.natRule, err = p.nat.MasqueradeSourcePort(p.listenPort, p.cp.Local.Port(), p.ep); err != nil {
				return err
			}

		case epdiscproto.ProxyType_KERNEL_CONN:
			// We cant to anything for prfx and relay candidates.
			// Let them pass through the userspace connection

			var create = false
			if p.connUser == nil {
				// We lazily create the user connection on demand to avoid opening unused sockets
				create = true
			} else if ra, ok := p.connUser.RemoteAddr().(*net.UDPAddr); ok && ra.Port != p.listenPort {
				// Also recreate the user connection in case the WireGuard listen port has changed
				create = true
			}

			var newConnUser *net.UDPConn
			if create {
				if newConnUser, err = p.newUserConn(newConnICE); err != nil {
					return fmt.Errorf("failed to setup user connection: %w", err)
				}
				p.logger.Info("Setup user-space proxy connection",
					zap.Any("localAddress", newConnUser.LocalAddr()),
					zap.Any("remoteAddress", newConnUser.RemoteAddr()))
			}

			// Start copying if the underlying ice.Conn has changed
			if newConnICE != p.connICE || newConnUser != p.connUser {
				if p.connUser != nil {
					if err := p.connUser.Close(); err != nil {
						return fmt.Errorf("failed to close old user connection: %w", err)
					}
				}

				p.connICE = newConnICE
				p.connUser = newConnUser

				// Bi-directional copy between ICE and loopback UDP sockets
				go p.copy(newConnICE, p.connUser)
				go p.copy(p.connUser, newConnICE)
			}

			p.ep = p.connUser.LocalAddr().(*net.UDPAddr)
		}
	}

	return nil
}

func (p *KernelProxy) copy(dst io.Writer, src io.Reader) {
	buf := make([]byte, maxSegmentSize)
	for {
		// TODO: Check why this is not working
		// if _, err := io.Copy(dst, src); err != nil {
		// 	p.logger.Error("Failed copy", zap.Error(err))
		// }

		n, err := src.Read(buf)
		if err != nil {
			if errors.Is(err, ice.ErrClosed) || errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}

			p.logger.Error("Failed to read", zap.Error(err))
			continue
		}

		if _, err = dst.Write(buf[:n]); err != nil {
			p.logger.Error("Failed to write", zap.Error(err))
		}
	}
}

func (p *KernelProxy) newUserConn(iceConn *ice.Conn) (*net.UDPConn, error) {
	// User-space proxying
	rAddr := net.UDPAddr{
		IP:   net.IPv6loopback,
		Port: int(p.listenPort),
	}

	lAddr := net.UDPAddr{
		IP:   net.IPv6loopback,
		Port: 0, // choose randomly
	}

	conn, err := net.DialUDP("udp", &lAddr, &rAddr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *KernelProxy) Type() epdiscproto.ProxyType {
	if p.cp == nil {
		return epdiscproto.ProxyType_NO_PROXY
	} else if p.cp.Local.Type() == ice.CandidateTypeHost || p.cp.Local.Type() == ice.CandidateTypeServerReflexive {
		return epdiscproto.ProxyType_KERNEL_NAT
	} else {
		return epdiscproto.ProxyType_KERNEL_CONN
	}
}
