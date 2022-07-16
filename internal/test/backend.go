package test

import (
	"context"
	"net/url"
	"strings"
	"sync"

	g "github.com/onsi/gomega"
	"riasc.eu/wice/internal/log"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
	"riasc.eu/wice/pkg/signaling"
)

type readyHandler struct {
	sync.WaitGroup
}

func (r *readyHandler) OnBackendReady(b signaling.Backend) {
	r.Done()
}

type peer struct {
	id       int64
	backend  signaling.Backend
	key      crypto.Key
	events   chan *pb.Event
	messages map[int64]chan *pb.SignalingMessage
}

func (p *peer) publish(o *peer) {
	if p.id == o.id {
		return
	}

	kp := &crypto.KeyPair{
		Ours:   p.key,
		Theirs: o.key.PublicKey(),
	}

	sentMsg := &pb.SignalingMessage{
		Session: &pb.SessionDescription{
			// We use the epoch to transport the id of the sending peer which gets checked on the receiving side
			// This should allow us to check against any mixed up message deliveries
			Epoch: p.id,
		},
	}

	err := p.backend.Publish(context.Background(), kp, sentMsg)
	g.Expect(err).To(g.Succeed(), "Failed to publish signaling message: %s", err)
}

func (p *peer) receive(o *peer) {
	recvMsg := <-p.messages[o.id]

	g.Expect(recvMsg.Session.Epoch).To(g.Equal(o.id), "Received invalid message")
}

// TestBackend creates n peers with separate connections to the signaling backend u
// and exchanges a test message between each pair of backends
func RunBackendTest(u string, n int) {
	// Add a colon to make url.Parse succeed
	if !strings.Contains(u, ":") {
		u += ":"
	}

	uri, err := url.Parse(u)
	g.Expect(err).To(g.Succeed(), "Failed to parse URL: %s", err)

	ready := &readyHandler{}
	ready.Add(n)

	cfg := &signaling.BackendConfig{
		URI: uri,
	}

	ps := []*peer{}
	for i := 0; i < n; i++ {
		p := &peer{
			id:       int64(i + 100),
			events:   log.NewEventLogger(),
			messages: map[int64]chan *pb.SignalingMessage{},
		}

		p.backend, err = signaling.NewBackend(cfg)
		g.Expect(err).To(g.Succeed(), "Failed to create backend: %s", err)

		defer p.backend.Close()

		p.key, err = crypto.GeneratePrivateKey()
		g.Expect(err).To(g.Succeed(), "Failed to generate private key: %s", err)

		ps = append(ps, p)
	}

	// Wait until all backends are ready
	ready.Wait()

	for _, p := range ps {
		for _, o := range ps {
			if p == o {
				continue // Do not send messages to ourself
			}

			kp := &crypto.KeyPair{
				Ours:   p.key,
				Theirs: o.key.PublicKey(),
			}

			p.messages[o.id], err = p.backend.Subscribe(context.Background(), kp)
			g.Expect(err).To(g.Succeed(), "Failed to subscribe: %s", err)
		}
	}

	wg := sync.WaitGroup{}
	wg.Add(n*n - n)

	for _, p := range ps {
		for _, o := range ps {
			if p.id == o.id {
				continue // Do not send messages to ourself
			}

			go func(p, o *peer) {
				p.receive(o)
				wg.Done()
			}(p, o)

			p.publish(o)
		}
	}

	wg.Wait()
}
