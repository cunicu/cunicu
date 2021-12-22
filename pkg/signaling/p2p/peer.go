package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/signaling"
)

type Peer struct {
	Offers chan signaling.Offer

	logger log.FieldLogger

	topic *pubsub.Topic
	sub   *pubsub.Subscription

	context context.Context
}

func (b *Backend) NewPeer(kp crypto.PublicKeyPair) (*Peer, error) {
	var err error

	p := &Peer{
		context: context.Background(),
		logger:  b.logger.WithField("kp", kp),
	}

	// topicFromPublicKeyPair derives a common topic name by XOR-ing the public keys of the peers
	// As Xor is a cummutative operation the topic name is the same from
	// both the viewpoint of both sides respectively.
	t := fmt.Sprintf("wice/pp/%s", kp.Shared())

	if p.topic, err = b.pubsub.Join(t); err != nil {
		return nil, err
	}

	if p.sub, err = p.topic.Subscribe(); err != nil {
		return nil, err
	}

	go p.readLoop()

	return p, err
}

func (p *Peer) publishOffer(o signaling.Offer) error {
	data, err := json.Marshal(&o)
	if err != nil {
		return err
	}

	go func() {
		for {
			time.Sleep(1 * time.Second)
			p.logger.WithFields(log.Fields{
				"offer": o,
				"topic": p.topic,
			}).Debug("Published offer to topic")
			p.topic.Publish(p.context, data)
		}
	}()
	return nil

	// return p.topic.Publish(p.context, data)
}

func (p *Peer) readLoop() {
	var err error
	var msg *pubsub.Message

	for {
		if msg, err = p.sub.Next(p.context); err != nil {
			close(p.Offers)
			return
		}

		// only forward messages delivered by others
		// if msg.ReceivedFrom == cr.self {
		// 	continue
		// }

		o := signaling.Offer{}
		if err := json.Unmarshal(msg.Data, &o); err != nil {
			p.logger.WithError(err).Error("Failed to decode received offer")
		}

		p.logger.WithFields(log.Fields{
			"offer": o,
			"topic": p.topic,
		}).Debug("Received offer")

		// send valid messages onto the Messages channel
		p.Offers <- o
	}
}
