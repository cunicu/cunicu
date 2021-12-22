package p2p

import (
	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
)

func (b *Backend) Listen(n network.Network, a ma.Multiaddr) {
	b.logger.WithField("addr", a).Debug("Started listening")
}

func (b *Backend) ListenClose(n network.Network, a ma.Multiaddr) {
	b.logger.WithField("addr", a).Debug("Stopped listening")
}

func (b *Backend) Connected(n network.Network, c network.Conn) {
	b.logger.WithField("remote", c.RemoteMultiaddr()).Debug("Connected")
}

func (b *Backend) Disconnected(n network.Network, c network.Conn) {
	b.logger.WithField("remote", c.RemoteMultiaddr()).Debug("Disconnected")

}

func (b *Backend) OpenedStream(n network.Network, s network.Stream) {
	b.logger.WithFields(logrus.Fields{
		"id":       s.ID(),
		"remote":   s.Conn().RemoteMultiaddr(),
		"protocol": s.Protocol(),
	}).Debug("Stream opened")

}

func (b *Backend) ClosedStream(n network.Network, s network.Stream) {
	b.logger.WithFields(logrus.Fields{
		"id":       s.ID(),
		"remote":   s.Conn().RemoteMultiaddr(),
		"protocol": s.Protocol(),
	}).Debug("Stream closed")
}
