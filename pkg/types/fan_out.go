// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"sync"
)

type FanOut[T any] struct {
	lock sync.RWMutex
	buf  int
	subs map[chan T]any
}

func NewFanOut[T any](buf int) *FanOut[T] {
	f := &FanOut[T]{
		subs: map[chan T]any{},
		buf:  buf,
	}

	return f
}

func (f *FanOut[T]) Send(v T) {
	f.lock.RLock()
	defer f.lock.RUnlock()

	for ch := range f.subs {
		ch <- v
	}
}

func (f *FanOut[T]) Add() chan T {
	ch := make(chan T, f.buf)

	f.lock.Lock()
	defer f.lock.Unlock()

	f.subs[ch] = nil

	return ch
}

func (f *FanOut[T]) Remove(ch chan T) {
	f.lock.Lock()
	defer f.lock.Unlock()

	delete(f.subs, ch)
}

func (f *FanOut[T]) Close() {
	f.lock.Lock()
	defer f.lock.Unlock()

	for ch := range f.subs {
		close(ch)
	}
}
