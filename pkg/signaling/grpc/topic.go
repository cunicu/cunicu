package grpc

import (
	"github.com/stv0g/cunicu/pkg/signaling"
	"github.com/stv0g/cunicu/pkg/util"
)

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
