package test

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	Messages map[crypto.Key]map[crypto.Key][]*pb.SignalingMessage
}

func NewMessageHandler() *msgHandler {
	return &msgHandler{
		Count:    atomic.NewUint32(0),
		Messages: map[crypto.Key]map[crypto.Key][]*pb.SignalingMessage{},
	}
}

func (h *msgHandler) OnSignalingMessage(kp *crypto.PublicKeyPair, msg *pb.SignalingMessage) {
	h.Count.Inc()

	if _, ok := h.Messages[kp.Ours]; !ok {
		h.Messages[kp.Ours] = map[crypto.Key][]*pb.SignalingMessage{}
	}

	if _, ok := h.Messages[kp.Ours][kp.Theirs]; !ok {
		h.Messages[kp.Ours][kp.Theirs] = []*pb.SignalingMessage{}
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
		if msg.Session.Epoch != int64(o.id) {
			return fmt.Errorf("received invalid msg: epoch == %d != %d", msg.Session.Epoch, o.id)
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

	sentMsg := &pb.SignalingMessage{
		Session: &pb.SessionDescription{
			// We use the epoch to transport the id of the sending peer which gets checked on the receiving side
			// This should allow us to check against any mixed up message deliveries
			Epoch: int64(p.id),
		},
	}

	return p.backend.Publish(context.Background(), kp, sentMsg)
}

// TestBackend creates n peers with separate connections to the signaling backend u
// and exchanges a test message between each pair of backends
func BackendTest(u *url.URL, n int) {
	var err error
	var ps []*peer

	BeforeEach(func() {
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
			Expect(err).To(Succeed(), "Failed to create backend: %s", err)

			p.key, err = crypto.GeneratePrivateKey()
			Expect(err).To(Succeed(), "Failed to generate private key: %s", err)

			ps = append(ps, p)
		}

		// Wait until all backends are ready
		Eventually(func() int { return int(backendReady.Count.Load()) }).Should(Equal(n))
	})

	AfterEach(func() {
		for _, p := range ps {
			err := p.backend.Close()
			Expect(err).To(Succeed())
		}
	})

	It("exchanges messages between multiple pairs", func() {
		mh1 := NewMessageHandler()
		mh2 := NewMessageHandler()
		mh3 := NewMessageHandler()
		mh4 := NewMessageHandler()

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
				Expect(err).To(Succeed())

				err = p.backend.Subscribe(context.Background(), kp, mh2)
				Expect(err).To(Succeed())
			}

			err = p.backend.SubscribeAll(context.Background(), &p.key, mh3)
			Expect(err).To(Succeed())

			err = p.backend.SubscribeAll(context.Background(), &p.key, mh4)
			Expect(err).To(Succeed())
		}

		// Send messages
		for _, p := range ps {
			for _, o := range ps {
				if p.id == o.id {
					continue // Do not send messages to ourself
				}

				err := p.publish(o)
				Expect(err).To(Succeed(), "Failed to publish signaling message: %s", err)
			}
		}

		// Wait until we have exchanged all messages
		Eventually(func() int { return int(mh1.Count.Load()) }).Should(BeNumerically(">=", n*n-n))
		Eventually(func() int { return int(mh2.Count.Load()) }).Should(BeNumerically(">=", n*n-n))
		Eventually(func() int { return int(mh3.Count.Load()) }).Should(BeNumerically(">=", n*n-n))
		Eventually(func() int { return int(mh4.Count.Load()) }).Should(BeNumerically(">=", n*n-n))

		// Check if we received the message
		for _, p := range ps {
			for _, o := range ps {
				if p.id == o.id {
					continue // Do not send messages to ourself
				}

				Expect(mh1.Check(p, o)).To(Succeed(), "Failed to receive message: %s", err)
				Expect(mh2.Check(p, o)).To(Succeed(), "Failed to receive message: %s", err)
				Expect(mh3.Check(p, o)).To(Succeed(), "Failed to receive message: %s", err)
				Expect(mh4.Check(p, o)).To(Succeed(), "Failed to receive message: %s", err)
			}
		}
	})
}
