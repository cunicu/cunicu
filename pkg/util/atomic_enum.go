package util

import "sync/atomic"

type AtomicEnum[T ~int] atomic.Uint64

func (a *AtomicEnum[T]) Load() T {
	return T((*atomic.Uint64)(a).Load())
}

func (a *AtomicEnum[T]) Store(v T) {
	(*atomic.Uint64)(a).Store(uint64(v))
}

func (a *AtomicEnum[T]) CompareAndSwap(oldVal, newVal T) bool {
	return (*atomic.Uint64)(a).CompareAndSwap(uint64(oldVal), uint64(newVal))
}

func (a *AtomicEnum[T]) Swap(newVal T) T {
	return T((*atomic.Uint64)(a).Swap(uint64(newVal)))
}