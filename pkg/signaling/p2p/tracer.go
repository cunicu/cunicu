package p2p

import (
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
)

type tracer struct {
	logger *zap.Logger
}

func newTracer(b *Backend) *tracer {
	return &tracer{
		logger: b.logger.Named("pubsub"),
	}
}

func (t *tracer) AddPeer(p peer.ID, proto protocol.ID) {
	t.logger.Debug("Peer added", zap.Any("peer", p))
}

func (t *tracer) RemovePeer(p peer.ID) {
	t.logger.Debug("Peer removed", zap.Any("peer", p))
}

func (t *tracer) Join(topic string) {
	t.logger.Debug("Topic joined", zap.String("topic", topic))
}

func (t *tracer) Leave(topic string) {
	t.logger.Debug("Topic lefft", zap.String("topic", topic))
}

func (t *tracer) Graft(p peer.ID, topic string) {
	t.logger.Debug("Graft peer to topic", zap.Any("peer", p), zap.String("topic", topic))
}

func (t *tracer) Prune(p peer.ID, topic string) {

}

func (t *tracer) ValidateMessage(msg *pubsub.Message) {

}

func (t *tracer) DeliverMessage(msg *pubsub.Message) {

}

func (t *tracer) RejectMessage(msg *pubsub.Message, reason string) {

}

func (t *tracer) DuplicateMessage(msg *pubsub.Message) {

}

func (t *tracer) ThrottlePeer(p peer.ID) {

}

func (t *tracer) RecvRPC(rpc *pubsub.RPC) {

}

func (t *tracer) SendRPC(rpc *pubsub.RPC, p peer.ID) {

}

func (t *tracer) DropRPC(rpc *pubsub.RPC, p peer.ID) {

}

func (t *tracer) UndeliverableMessage(msg *pubsub.Message) {

}
