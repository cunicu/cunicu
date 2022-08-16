package test

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	gi "github.com/onsi/ginkgo/v2"
	g "github.com/onsi/gomega"
	"go.uber.org/atomic"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

type readyHandler struct {
	Count *atomic.Uint32
}

func (r *readyHandler) OnSignalingBackendReady(b signaling.Backend) {
	r.Count.Inc()
}

type msgHandler struct {
	Count    *atomic.Uint32
	Messages map[crypto.Key]map[crypto.Key][]*signaling.Message
}

func newMessageHandler() *msgHandler {
	return &msgHandler{
		Count:    atomic.NewUint32(0),
		Messages: map[crypto.Key]map[crypto.Key][]*signaling.Message{},
	}
}

func (h *msgHandler) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *signaling.Message) {
	h.Count.Inc()

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

	if found > 1 {
		return fmt.Errorf("peer %d received %d messages from peer %d", p.id, found, o.id)
	} else if found == 0 {
		return fmt.Errorf("peer %d received no messages from peer %d", p.id, o.id)
	} else {
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
		Candidate: &pb.Candidate{
			// We use the epoch to transport the id of the sending peer which gets checked on the receiving side
			// This should allow us to check against any mixed up message deliveries
			Port: int32(p.id),
		},
	}

	return p.backend.Publish(context.Background(), kp, sentMsg)
}

// TestBackend creates n peers with separate connections to the signaling backend u
// and exchanges a test message between each pair of backends
func BackendTest(u *url.URL, n int) {
	var err error
	var ps []*peer

	gi.BeforeEach(func() {
		backendReady := &readyHandler{
			Count: atomic.NewUint32(0),
		}

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
			g.Expect(err).To(g.Succeed(), "Failed to create backend: %s", err)

			p.key, err = crypto.GeneratePrivateKey()
			g.Expect(err).To(g.Succeed(), "Failed to generate private key: %s", err)

			ps = append(ps, p)
		}

		// Wait until all backends are ready
		g.Eventually(func() int { return int(backendReady.Count.Load()) }).Should(g.Equal(n))
	})

	gi.AfterEach(func() {
		for _, p := range ps {
			err := p.backend.Close()
			g.Expect(err).To(g.Succeed())
		}
	})

	gi.It("exchanges messages between multiple pairs", func() {
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

				err = p.backend.Subscribe(context.Background(), kp, mh1)
				g.Expect(err).To(g.Succeed())

				err = p.backend.Subscribe(context.Background(), kp, mh2)
				g.Expect(err).To(g.Succeed())
			}

			err = p.backend.SubscribeAll(context.Background(), &p.key, mh3)
			g.Expect(err).To(g.Succeed())

			err = p.backend.SubscribeAll(context.Background(), &p.key, mh4)
			g.Expect(err).To(g.Succeed())
		}

		// Send messages
		for _, p := range ps {
			for _, o := range ps {
				if p.id == o.id {
					continue // Do not send messages to ourself
				}

				err := p.publish(o)
				g.Expect(err).To(g.Succeed(), "Failed to publish signaling message: %s", err)
			}
		}

		// Wait until we have exchanged all messages
		g.Eventually(func() int { return int(mh1.Count.Load()) }).Should(g.BeNumerically(">=", n*n-n))
		g.Eventually(func() int { return int(mh2.Count.Load()) }).Should(g.BeNumerically(">=", n*n-n))
		g.Eventually(func() int { return int(mh3.Count.Load()) }).Should(g.BeNumerically(">=", n*n-n))
		g.Eventually(func() int { return int(mh4.Count.Load()) }).Should(g.BeNumerically(">=", n*n-n))

		// Check if we received the message
		for _, p := range ps {
			for _, o := range ps {
				if p.id == o.id {
					continue // Do not send messages to ourself
				}

				g.Expect(mh1.Check(p, o)).To(g.Succeed(), "Failed to receive message: %s", err)
				g.Expect(mh2.Check(p, o)).To(g.Succeed(), "Failed to receive message: %s", err)
				g.Expect(mh3.Check(p, o)).To(g.Succeed(), "Failed to receive message: %s", err)
				g.Expect(mh4.Check(p, o)).To(g.Succeed(), "Failed to receive message: %s", err)
			}
		}
	})
}
