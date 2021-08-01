package base

import (
	"net/url"

	log "github.com/sirupsen/logrus"
	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/crypto"
)

type Backend struct {
	Offers map[crypto.PublicKeyPair]chan backend.Offer

	Logger log.FieldLogger
	Type   string
}

func NewBackend(uri *url.URL, options map[string]string) Backend {
	logFields := log.Fields{
		"logger":  "backend",
		"backend": uri.Scheme,
	}

	b := Backend{
		Offers: make(map[crypto.PublicKeyPair]chan backend.Offer),
		Logger: log.WithFields(logFields),
	}

	return b
}

func (b *Backend) Close() error {
	return nil
}

func (b *Backend) SubscribeOffers(kp crypto.PublicKeyPair) chan backend.Offer {
	b.Logger.WithField("kp", kp).Info("Subscribe to offers from peer")

	ch, ok := b.Offers[kp]
	if !ok {
		ch = make(chan backend.Offer, 100)
		b.Offers[kp] = ch
	}

	return ch
}

func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer backend.Offer) error {
	b.Logger.WithField("kp", kp).WithField("offer", offer).Debug("Published offer")

	return nil
}

func (b *Backend) WithdrawOffer(kp crypto.PublicKeyPair) error {
	b.Logger.WithField("kp", kp).Debug("Withdrawed offer")

	return nil
}
