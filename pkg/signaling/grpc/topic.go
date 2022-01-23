package grpc

import (
	"sync"

	"riasc.eu/wice/pkg/pb"
)

type topic struct {
	subs     map[chan *pb.Offer]bool
	subsLock sync.RWMutex
}

func newTopic() *topic {
	return &topic{
		subs: make(map[chan *pb.Offer]bool),
	}
}

func (t *topic) Publish(o *pb.Offer) {
	t.subsLock.RLock()
	defer t.subsLock.RUnlock()

	for s := range t.subs {
		s <- o
	}
}

func (t *topic) Subscribe() chan *pb.Offer {
	t.subsLock.Lock()
	defer t.subsLock.Unlock()

	c := make(chan *pb.Offer)

	t.subs[c] = true

	return c
}

func (t *topic) Unsubscribe(ch chan *pb.Offer) {
	t.subsLock.Lock()
	defer t.subsLock.Unlock()

	close(ch)
	delete(t.subs, ch)
}
