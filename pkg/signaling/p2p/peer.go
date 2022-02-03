package p2p

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p-core/host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type Peer struct {
	Messages chan *pb.SignalingMessage

	keyPair *crypto.KeyPair
	host    host.Host
	topic   *pubsub.Topic
	sub     *pubsub.Subscription

	logger *zap.Logger

	context context.Context
}

func (b *Backend) NewPeer(kp *crypto.KeyPair) (*Peer, error) {
	var err error

	p := &Peer{
		keyPair:  kp,
		host:     b.host,
		context:  context.Background(),
		logger:   b.logger.With(zap.Any("kp", kp)),
		Messages: make(chan *pb.SignalingMessage, 100),
	}

	// topicFromPublicKeyPair derives a common topic name by XOR-ing the public keys of the peers
	// As Xor is a cummutative operation the topic name is the same from
	// both the viewpoint of both sides respectively.
	t := fmt.Sprintf("wice/pp/%s", kp.Shared())

	if p.topic, err = b.pubsub.Join(t); err != nil {
		return nil, fmt.Errorf("failed to join topic: %w", err)
	}

	if p.sub, err = p.topic.Subscribe(); err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	go p.readLoop()

	return p, nil
}

func (p *Peer) publishMessage(offer *pb.SignalingMessage) error {
	payload, err := proto.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	if err := p.topic.Publish(p.context, payload, pubsub.WithReadiness(pubsub.MinTopicSize(1))); err != nil {
		return fmt.Errorf("failed to publish offer: %w", err)
	}

	p.topic.Relay()

	return nil
}

func (p *Peer) readLoop() {
	var err error
	var msg *pubsub.Message

	for {
		if msg, err = p.sub.Next(p.context); err != nil {
			p.logger.Error("Failed to receive offers", zap.Error(err))
			close(p.Messages)
			return
		}

		if msg.ReceivedFrom == p.host.ID() {
			continue
		}

		env := &pb.SignalingEnvelope{}
		if err := proto.Unmarshal(msg.Data, env); err != nil {
			p.logger.Error("Failed to unmarshal offer", zap.Error(err))
			continue
		}

		msg := &pb.SignalingMessage{}
		if err := env.Contents.Unmarshal(msg, p.keyPair); err != nil {
			p.logger.Error("Failed to decrypt message", zap.Error(err))
		}

		// send valid messages onto the Messages channel
		p.Messages <- msg
	}
}
