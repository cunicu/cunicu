package p2p

import (
	"context"
	"encoding/json"
	"fmt"

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

	t := topicFromPublicKeyPair(kp)

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

	p.logger.WithFields(log.Fields{
		"offer": o,
		"topic": p.topic,
	}).Debug("Published offer to topic")

	return p.topic.Publish(p.context, data)
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

		p.logger.WithField("offer", o).Debug("Received offer")

		// send valid messages onto the Messages channel
		p.Offers <- o
	}
}

func topicFromPublicKeyPair(kp crypto.PublicKeyPair) string {
	return fmt.Sprintf("wice/%s/%s", kp.Ours, kp.Theirs)
}
