package proxy

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/pion/ice/v2"
	"go.uber.org/zap"

	"github.com/stv0g/cunicu/pkg/wg"

	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
)

type UserBindProxy struct {
	bind     *wg.UserBind
	endpoint *wg.UserEndpoint

	conn *ice.Conn

	logger *zap.Logger
}

func NewUserBindProxy(bind *wg.UserBind) (*UserBindProxy, error) {
	return &UserBindProxy{
		bind:   bind,
		logger: zap.L().Named("proxy").With(zap.String("type", "user-bind")),
	}, nil
}

func (p *UserBindProxy) Close() error {
	return nil
}

func (p *UserBindProxy) UpdateCandidatePair(cp *ice.CandidatePair, conn *ice.Conn) (*net.UDPAddr, error) {
	var err error

	p.logger.Debug("Forwarding via in-process bind")

	ep := &net.UDPAddr{
		IP:   net.ParseIP(cp.Remote.Address()),
		Port: cp.Remote.Port(),
	}

	p.endpoint, err = p.bind.UpdateEndpoint(ep, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to update endpoint: %w", err)
	}

	if conn != p.conn {
		go p.read(conn)
	}

	return ep, nil
}

func (p *UserBindProxy) read(conn *ice.Conn) {
	p.conn = conn

	for {
		buf := make([]byte, maxSegmentSize)

		n, err := p.conn.Read(buf)
		if err != nil {
			if errors.Is(err, ice.ErrClosed) || errors.Is(err, io.EOF) {
				return
			}

			p.logger.Error("Failed to read from ICE connection", zap.Error(err))
			continue
		}

		if err := p.bind.OnData(buf[:n], p.endpoint); err != nil {
			p.logger.Error("Failed to pass data to bind", zap.Error(err))
		}
	}
}

func (p *UserBindProxy) Type() epdiscproto.ProxyType {
	return epdiscproto.ProxyType_USER_BIND
}
