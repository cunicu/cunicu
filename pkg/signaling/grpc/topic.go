package grpc

import (
	"sync"

	"riasc.eu/wice/pkg/crypto"
	"riasc.eu/wice/pkg/signaling"
	"riasc.eu/wice/pkg/util"
)

type topicRegistry struct {
	topics     map[crypto.Key]*topic
	topicsLock sync.Mutex
}

func (r *topicRegistry) getTopic(pk *crypto.Key) *topic {
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

type topic struct {
	subs *util.FanOut[*signaling.Envelope]
}

func NewTopic() *topic {
	t := &topic{
		subs: util.NewFanOut[*signaling.Envelope](128),
	}

	return t
}

func (t *topic) Publish(env *signaling.Envelope) {
	t.subs.Send(env)
}

func (t *topic) Subscribe() chan *signaling.Envelope {
	return t.subs.Add()
}

func (t *topic) Unsubscribe(ch chan *signaling.Envelope) {
	t.subs.Remove(ch)
}

func (t *topic) Close() {
	t.subs.Close()
}
