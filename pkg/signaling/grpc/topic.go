package grpc

import (
	"sync"

	"go.uber.org/zap"
	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/pb"
)

type topicRegistry struct {
	topics     map[crypto.Key]*topic
	topicsLock sync.Mutex
}

func (s *topicRegistry) getTopic(pk *crypto.Key) *topic {
	s.topicsLock.Lock()
	defer s.topicsLock.Unlock()

	top, ok := s.topics[*pk]
	if ok {
		return top
	}

	top = newTopic()

	s.topics[*pk] = top

	return top
}

type topic struct {
	subs     map[chan *pb.SignalingEnvelope]bool
	subsLock sync.RWMutex
	subsCond *sync.Cond
}

func newTopic() *topic {
	t := &topic{
		subs: make(map[chan *pb.SignalingEnvelope]bool),
	}

	t.subsCond = sync.NewCond(&t.subsLock)

	return t
}

func (t *topic) Publish(env *pb.SignalingEnvelope) {
	t.subsLock.RLock()
	defer t.subsLock.RUnlock()

	for s := range t.subs {
		s <- env
	}
}

func (t *topic) Subscribe() chan *pb.SignalingEnvelope {
	t.subsLock.Lock()
	defer t.subsLock.Unlock()

	c := make(chan *pb.SignalingEnvelope)

	t.subs[c] = true

	t.subsCond.Broadcast()

	return c
}

func (t *topic) Unsubscribe(ch chan *pb.SignalingEnvelope) {
	t.subsLock.Lock()
	defer t.subsLock.Unlock()

	close(ch)
	delete(t.subs, ch)
}

func (t *topic) WaitForSubs(num int) {
	t.subsLock.Lock()
	defer t.subsLock.Unlock()

	for len(t.subs) < num {
		zap.L().Info("Wait for subs", zap.Int("have", len(t.subs)), zap.Int("want", num))
		t.subsCond.Wait()
	}
}
