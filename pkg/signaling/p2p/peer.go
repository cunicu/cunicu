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
	Offers chan *pb.Offer

	host  host.Host
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	logger *zap.Logger

	context context.Context
}

func (b *Backend) NewPeer(kp crypto.PublicKeyPair) (*Peer, error) {
	var err error

	p := &Peer{
		host:    b.host,
		context: context.Background(),
		logger:  b.logger.With(zap.Any("kp", kp)),
		Offers:  make(chan *pb.Offer, 100),
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

	p.logger.Debug("Starting reading messages", zap.String("topic", t))

	go p.readLoop()

	return p, nil
}

func (p *Peer) publishOffer(offer *pb.Offer) error {
	payload, err := proto.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to marshal offer: %w", err)
	}

	p.logger.Debug("Publishing offer to topic",
		zap.Any("offer", offer),
		zap.Any("topic", p.topic),
	)

	return p.topic.Publish(p.context, payload)
}

func (p *Peer) readLoop() {
	var err error
	var msg *pubsub.Message

	for {
		if msg, err = p.sub.Next(p.context); err != nil {
			p.logger.Error("Failed to receive offers", zap.Error(err))
			close(p.Offers)
			return
		}

		if msg.ReceivedFrom == p.host.ID() {
			continue
		}

		offer := &pb.Offer{}
		if err := proto.Unmarshal(msg.Data, offer); err != nil {
			p.logger.Error("Failed to unmarshal offer", zap.Error(err))
			continue
		}

		p.logger.Debug("Received offer",
			zap.Any("offer", offer),
			zap.Any("topic", p.topic),
			zap.Any("from", msg.ReceivedFrom),
		)

		// send valid messages onto the Messages channel
		p.Offers <- offer
	}
}
