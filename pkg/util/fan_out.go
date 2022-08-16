package util

import "sync"

type FanOut[T any] struct {
	C chan T

	buf  int
	subs map[chan T]struct{}
	lock sync.RWMutex
}

func NewFanOut[T any](buf int) *FanOut[T] {
	f := &FanOut[T]{
		C:    make(chan T),
		subs: map[chan T]struct{}{},
		buf:  buf,
	}

	go f.run()

	return f
}

func (f *FanOut[T]) run() {
	for t := range f.C {
		f.lock.RLock()
		for ch := range f.subs {
			ch <- t
		}
		f.lock.RUnlock()
	}
}

func (f *FanOut[T]) Add() chan T {
	ch := make(chan T, f.buf)

	f.lock.Lock()
	defer f.lock.Unlock()

	f.subs[ch] = struct{}{}

	return ch
}

func (f *FanOut[T]) Remove(ch chan T) {
	f.lock.Lock()
	defer f.lock.Unlock()

	delete(f.subs, ch)
}

func (f *FanOut[T]) Close() error {
	f.lock.Lock()
	defer f.lock.Unlock()

	for ch := range f.subs {
		close(ch)
	}

	close(f.C)

	return nil
}
