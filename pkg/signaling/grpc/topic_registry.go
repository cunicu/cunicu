package grpc

import (
	"sync"

	"github.com/stv0g/cunicu/pkg/crypto"
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
