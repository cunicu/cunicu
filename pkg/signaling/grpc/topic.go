package grpc

import (
	"sync"

	"github.com/stv0g/cunicu/pkg/crypto"
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"
)

type topicRegistry struct {
	topics     map[crypto.Key]*Topic
	topicsLock sync.Mutex
}

func (r *topicRegistry) getTopic(pk *crypto.Key) *Topic {
	r.topicsLock.Lock()
	defer r.topicsLock.Unlock()

	top, ok := r.topics[*pk]
	if ok {
		return top
	}

	top = NewTopic()

	r.topics[*pk] = top

	return top
}

func (r *topicRegistry) Close() error {
	r.topicsLock.Lock()
	defer r.topicsLock.Unlock()

	for _, t := range r.topics {
		t.Close()
	}

	return nil
}

type Topic struct {
	subs *util.FanOut[*signaling.Envelope]
}

func NewTopic() *Topic {
	t := &Topic{
		subs: util.NewFanOut[*signaling.Envelope](128),
	}

	return t
}

func (t *Topic) Publish(env *signaling.Envelope) {
	t.subs.Send(env)
}

func (t *Topic) Subscribe() chan *signaling.Envelope {
	return t.subs.Add()
}

func (t *Topic) Unsubscribe(ch chan *signaling.Envelope) {
	t.subs.Remove(ch)
}

func (t *Topic) Close() {
	t.subs.Close()
}
