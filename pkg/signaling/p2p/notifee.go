package p2p

import (
	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

func (b *Backend) Listen(n network.Network, a ma.Multiaddr) {
	b.logger.Debug("Started listening", zap.Any("addr", a))
}

func (b *Backend) ListenClose(n network.Network, a ma.Multiaddr) {
	b.logger.Debug("Stopped listening", zap.Any("addr", a))
}

func (b *Backend) Connected(n network.Network, c network.Conn) {
	b.logger.Debug("Connected", zap.Any("remote", c.RemoteMultiaddr()))
}

func (b *Backend) Disconnected(n network.Network, c network.Conn) {
	b.logger.Debug("Disconnected", zap.Any("remote", c.RemoteMultiaddr()))

}

func (b *Backend) OpenedStream(n network.Network, s network.Stream) {
	b.logger.Debug("Stream opened",
		zap.Any("id", s.ID()),
		zap.Any("remote", s.Conn().RemoteMultiaddr()),
		zap.Any("protocol", s.Protocol()),
	)

}

func (b *Backend) ClosedStream(n network.Network, s network.Stream) {
	b.logger.Debug("Stream closed",
		zap.Any("id", s.ID()),
		zap.Any("remote", s.Conn().RemoteMultiaddr()),
		zap.Any("protocol", s.Protocol()),
	)
}
