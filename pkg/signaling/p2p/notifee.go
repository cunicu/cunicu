package p2p

import (
	"github.com/libp2p/go-libp2p-core/network"
	ma "github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

type notifee struct {
	logger *zap.Logger
}

func newNotifee(b *Backend) *notifee {
	return &notifee{
		logger: b.logger.Named("p2p"),
	}
}

func (n *notifee) Listen(_ network.Network, a ma.Multiaddr) {
	n.logger.Debug("Started listening on address", zap.Any("addr", a))
}

func (n *notifee) ListenClose(_ network.Network, a ma.Multiaddr) {
	n.logger.Debug("Stopped listening on address", zap.Any("addr", a))
}

func (n *notifee) Connected(_ network.Network, c network.Conn) {
	n.logger.Debug("Connected to remote", zap.Any("remote", c.RemoteMultiaddr()))
}

func (n *notifee) Disconnected(_ network.Network, c network.Conn) {
	n.logger.Debug("Disconnected from remote", zap.Any("remote", c.RemoteMultiaddr()))
}

func (n *notifee) OpenedStream(_ network.Network, s network.Stream) {
	// n.logger.Debug("Stream opened",
	// 	zap.Any("id", s.ID()),
	// 	zap.Any("remote", s.Conn().RemoteMultiaddr()),
	// 	zap.Any("protocol", s.Protocol()),
	// )
}

func (n *notifee) ClosedStream(_ network.Network, s network.Stream) {
	// n.logger.Debug("Stream closed",
	// 	zap.Any("id", s.ID()),
	// 	zap.Any("remote", s.Conn().RemoteMultiaddr()),
	// 	zap.Any("protocol", s.Protocol()),
	// )
}
