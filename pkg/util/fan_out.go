package util

import "sync"

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
	f.lock.Lock()
	defer f.lock.Unlock()

	ch := make(chan T, f.buf)

	f.subs[ch] = nil

	return ch
}

func (f *FanOut[T]) Remove(ch chan T) {
	f.lock.Lock()
	defer f.lock.Unlock()

	delete(f.subs, ch)
}

func (f *FanOut[T]) Close() {
	for ch := range f.subs {
		close(ch)
	}
}
