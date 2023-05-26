// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"sync/atomic"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

type AtomicEnum[T constraints.Integer] atomic.Uint64

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

// SetIf updates the atomic value if the previous value
// matches one of the supplied values.
// It returns true if the value has been changed.
func (a *AtomicEnum[T]) SetIf(newValue T, prevValue ...T) (T, bool) {
	for {
		curValue := a.Load()

		if !slices.Contains(prevValue, curValue) {
			return curValue, false
		}

		if a.CompareAndSwap(curValue, newValue) {
			return curValue, true
		}
	}
}

// SetIfNot updates the atomic value if the previous value
// does not match any of the supplied values.
func (a *AtomicEnum[T]) SetIfNot(newValue T, prevValue ...T) (T, bool) {
	for {
		curValue := a.Load()

		if slices.Contains(prevValue, curValue) {
			return curValue, false
		}

		if a.CompareAndSwap(curValue, newValue) {
			return curValue, true
		}
	}
}
