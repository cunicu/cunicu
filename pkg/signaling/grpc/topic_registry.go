// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"sync"

	"cunicu.li/cunicu/pkg/crypto"
)

type topicRegistry struct {
	topics     map[crypto.Key]*Topic
	topicsLock sync.RWMutex
}

func (r *topicRegistry) getTopic(pk *crypto.Key) *Topic {
	r.topicsLock.RLock()
	top, ok := r.topics[*pk]
	r.topicsLock.RUnlock()

	if ok {
		return top
	}

	top = NewTopic()

	r.topicsLock.Lock()
	r.topics[*pk] = top
	r.topicsLock.Unlock()

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
