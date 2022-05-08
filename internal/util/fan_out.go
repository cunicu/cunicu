package util

import "sync"

type Fanout[T any] struct {
	C chan T

	buf  int
	subs map[chan T]struct{}
	lock sync.RWMutex
}

func NewFanout[T any](buf int) *Fanout[T] {
	f := &Fanout[T]{
		C:    make(chan T),
		subs: map[chan T]struct{}{},
		buf:  buf,
	}

	go f.run()

	return f
}

func (f *Fanout[T]) run() {
	for t := range f.C {
		f.lock.RLock()
		for ch := range f.subs {
			ch <- t
		}
		f.lock.RUnlock()
	}
}

func (f *Fanout[T]) Add() chan T {
	ch := make(chan T, f.buf)

	f.lock.Lock()
	f.subs[ch] = struct{}{}
	f.lock.Unlock()

	return ch
}

func (f *Fanout[T]) Remove(ch chan T) {
	f.lock.Lock()
	delete(f.subs, ch)
	f.lock.Unlock()
}

func (f *Fanout[T]) Close() error {
	close(f.C)
	return nil
}
