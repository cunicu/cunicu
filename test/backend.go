// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/stv0g/cunicu/pkg/crypto"
	epdiscproto "github.com/stv0g/cunicu/pkg/proto/feature/epdisc"
	"github.com/stv0g/cunicu/pkg/signaling"
)

type readyHandler struct {
	Count atomic.Uint32
}

func (r *readyHandler) OnSignalingBackendReady(_ signaling.Backend) {
	r.Count.Add(1)
}

type msgHandler struct {
	Count    atomic.Uint32
	Messages map[crypto.Key]map[crypto.Key][]*signaling.Message
}

func newMessageHandler() *msgHandler {
	return &msgHandler{
		Messages: map[crypto.Key]map[crypto.Key][]*signaling.Message{},
	}
}

func (h *msgHandler) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	h.Count.Add(1)

	if _, ok := h.Messages[kp.Ours]; !ok {
		h.Messages[kp.Ours] = map[crypto.Key][]*signaling.Message{}
	}

	if _, ok := h.Messages[kp.Ours][kp.Theirs]; !ok {
		h.Messages[kp.Ours][kp.Theirs] = []*signaling.Message{}
	}

	h.Messages[kp.Ours][kp.Theirs] = append(h.Messages[kp.Ours][kp.Theirs], msg)
}

func (h *msgHandler) Check(p, o *peer) error {
	kp := crypto.PublicKeyPair{
		Ours:   p.key.PublicKey(),
		Theirs: o.key.PublicKey(),
	}

	msgs, ok := h.Messages[kp.Ours]
	if !ok {
		return errors.New("did not find our key")
	}

	msgs2, ok := msgs[kp.Theirs]
	if !ok {
		return errors.New("did not find their key")
	}

	found := len(msgs2)

	switch {
	case found > 1:
		return fmt.Errorf("peer %d received %d messages from peer %d", p.id, found, o.id)
	case found == 0:
		return fmt.Errorf("peer %d received no messages from peer %d", p.id, o.id)
	default:
		msg := msgs2[0]
		if msg.Candidate.Port != int32(o.id) {
			return fmt.Errorf("received invalid msg: epoch == %d != %d", msg.Candidate.Port, o.id)
		}

		return nil
	}
}

type peer struct {
	id      int
	backend signaling.Backend
	key     crypto.Key
}

func (p *peer) publish(o *peer) error {
	kp := &crypto.KeyPair{
		Ours:   p.key,
		Theirs: o.key.PublicKey(),
	}

	sentMsg := &signaling.Message{
		Candidate: &epdiscproto.Candidate{
			// We use the epoch to transport the id of the sending peer which gets checked on the receiving side
			// This should allow us to check against any mixed up message deliveries
			Port: int32(p.id),
		},
	}

	return p.backend.Publish(context.Background(), kp, sentMsg)
}

// TestBackend creates n peers with separate connections to the signaling backend u
// and exchanges a test message between each pair of backends
func BackendTest(u *url.URL, n int) { //nolint:gocognit
	var err error
	var ps []*peer

	ginkgo.BeforeEach(func() {
		backendReady := &readyHandler{}

		ps = []*peer{}
		for i := 0; i < n; i++ {
			p := &peer{
				id: i + 100,
			}

			cfg := &signaling.BackendConfig{
				URI:     u,
				OnReady: []signaling.BackendReadyHandler{backendReady},
			}

			p.backend, err = signaling.NewBackend(cfg)
			gomega.Expect(err).To(gomega.Succeed(), "Failed to create backend: %s", err)

			p.key, err = crypto.GeneratePrivateKey()
			gomega.Expect(err).To(gomega.Succeed(), "Failed to generate private key: %s", err)

			ps = append(ps, p)
		}

		// Wait until all backends are ready
		gomega.Eventually(func() int { return int(backendReady.Count.Load()) }).Should(gomega.Equal(n))
	})

	ginkgo.AfterEach(func() {
		for _, p := range ps {
			err := p.backend.Close()
			gomega.Expect(err).To(gomega.Succeed())
		}
	})

	ginkgo.It("exchanges messages between multiple pairs", func() {
		mh1 := newMessageHandler()
		mh2 := newMessageHandler()
		mh3 := newMessageHandler()
		mh4 := newMessageHandler()

		// Subscribe
		for _, p := range ps {
			for _, o := range ps {
				if p == o {
					continue // Do not send messages to ourself
				}

				kp := &crypto.KeyPair{
					Ours:   p.key,
					Theirs: o.key.PublicKey(),
				}

				_, err = p.backend.Subscribe(context.Background(), kp, mh1)
				gomega.Expect(err).To(gomega.Succeed())

				_, err = p.backend.Subscribe(context.Background(), kp, mh2)
				gomega.Expect(err).To(gomega.Succeed())
			}

			kp := &crypto.KeyPair{
				Ours:   p.key,
				Theirs: crypto.Key{},
			}

			_, err = p.backend.Subscribe(context.Background(), kp, mh3)
			gomega.Expect(err).To(gomega.Succeed())

			_, err = p.backend.Subscribe(context.Background(), kp, mh4)
			gomega.Expect(err).To(gomega.Succeed())
		}

		// Send messages
		for _, p := range ps {
			for _, o := range ps {
				if p.id == o.id {
					continue // Do not send messages to ourself
				}

				err := p.publish(o)
				gomega.Expect(err).To(gomega.Succeed(), "Failed to publish signaling message: %s", err)
			}
		}

		// Wait until we have exchanged all messages
		gomega.Eventually(func() int { return int(mh1.Count.Load()) }).Should(gomega.BeNumerically(">=", n*n-n))
		gomega.Eventually(func() int { return int(mh2.Count.Load()) }).Should(gomega.BeNumerically(">=", n*n-n))
		gomega.Eventually(func() int { return int(mh3.Count.Load()) }).Should(gomega.BeNumerically(">=", n*n-n))
		gomega.Eventually(func() int { return int(mh4.Count.Load()) }).Should(gomega.BeNumerically(">=", n*n-n))

		// Check if we received the message
		for _, p := range ps {
			for _, o := range ps {
				if p.id == o.id {
					continue // Do not send messages to ourself
				}

				gomega.Expect(mh1.Check(p, o)).To(gomega.Succeed(), "Failed to receive message: %s", err)
				gomega.Expect(mh2.Check(p, o)).To(gomega.Succeed(), "Failed to receive message: %s", err)
				gomega.Expect(mh3.Check(p, o)).To(gomega.Succeed(), "Failed to receive message: %s", err)
				gomega.Expect(mh4.Check(p, o)).To(gomega.Succeed(), "Failed to receive message: %s", err)
			}
		}
	})
}
