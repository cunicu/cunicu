package http

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"riasc.eu/wice/pkg/backend"
	"riasc.eu/wice/pkg/backend/base"
	"riasc.eu/wice/pkg/crypto"
)

type Backend struct {
	base.Backend
	config BackendConfig

	client *http.Client
}

func init() {
	p := backend.BackendPlugin{
		New:         NewBackend,
		Description: "Simple HTTP/HTTPs REST API server",
	}

	backend.Backends["http"] = &p
	backend.Backends["https"] = &p
}

func NewBackend(uri *url.URL, options map[string]string) (backend.Backend, error) {
	b := &Backend{
		Backend: base.NewBackend(uri, options),
	}

	b.config.Parse(uri, options)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: b.config.InsecureSkipVerify,
		},
	}
	b.client = &http.Client{
		Transport: tr,
		Timeout:   b.config.Timeout,
	}

	go b.pollOffers()

	return b, nil
}

func (b *Backend) SubscribeOffer(kp crypto.PublicKeyPair) (chan backend.Offer, error) {
	ch := b.Backend.SubscribeOffers(kp)

	// Get initial offer without waiting for poller
	o, err := b.getOffer(kp)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer: %w", err)
	}

	if o.ID != 0 {
		ch <- o
	}

	return ch, nil
}

// pollOffers periodically fetches offers from the HTTP API and feeds them into the subscribption channels
func (b *Backend) pollOffers() {
	b.Logger.Info("Start polling for new offers")

	ticker := time.NewTicker(b.config.PollInterval)
	for range ticker.C {
		for kp, ch := range b.Offers {
			o, err := b.getOffer(kp)
			if err != nil {
				b.Logger.WithError(err).Error("Failed to fetch offer")
				continue
			}

			if o.ID != 0 {
				ch <- o
			}
		}
	}
}

// PublishOffer POSTs the Offer to the HTTP API
func (b *Backend) PublishOffer(kp crypto.PublicKeyPair, offer backend.Offer) error {
	buf, err := json.Marshal(offer)
	if err != nil {
		return fmt.Errorf("failed to encode offer: %w", err)
	}

	resp, err := b.client.Post(b.offerUrl(kp, false), "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("failed HTTP request: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed HTTP request: %s", resp.Status)
	}

	return b.Backend.PublishOffer(kp, offer)
}

func (b *Backend) getOffer(kp crypto.PublicKeyPair) (backend.Offer, error) {

	b.Logger.WithField("kp", kp).Trace("Fetching offer")

	resp, err := b.client.Get(b.offerUrl(kp, true))
	if err != nil {
		return backend.Offer{}, fmt.Errorf("failed HTTP request: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return backend.Offer{}, nil
		} else {
			return backend.Offer{}, fmt.Errorf("failed HTTP request: %s", resp.Status)
		}
	}

	var offer backend.Offer
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&offer)
	if err != nil {
		return backend.Offer{}, err
	}

	b.Logger.WithField("offer", offer).Debug("Fetched offer")

	return offer, nil
}

func (b *Backend) WithdrawOffer(kp crypto.PublicKeyPair) error {

	req, err := http.NewRequest(http.MethodDelete, b.offerUrl(kp, false), nil)
	if err != nil {
		return err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed HTTP request: %w", err)
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed HTTP request: %s", resp.Status)
	}

	return b.Backend.WithdrawOffer(kp)
}

func (b *Backend) Close() error {
	b.client.CloseIdleConnections()

	return nil // TODO
}

func (b *Backend) offerUrl(kp crypto.PublicKeyPair, sub bool) string {
	u := *b.config.URI
	if sub {
		u.Path += "/offers/" + url.PathEscape(kp.Theirs.String()) + "/" + url.PathEscape(kp.Ours.String())
	} else {
		u.Path += "/offers/" + url.PathEscape(kp.Ours.String()) + "/" + url.PathEscape(kp.Theirs.String())
	}
	return u.String()
}
