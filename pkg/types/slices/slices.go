// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package slices

import (
	"fmt"
	"math/rand"

	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

func Diff[T constraints.Ordered](oldSlice, newSlice []T) (added, removed, kept []T) {
	return DiffFunc(oldSlice, newSlice, func(a, b T) int {
		switch {
		case a == b:
			return 0
		case a < b:
			return -1
		default:
			return 1
		}
	})
}

func DiffFunc[T any](oldSlice, newSlice []T, cmp func(a, b T) int) (added, removed, kept []T) {
	added = []T{}
	removed = []T{}
	kept = []T{}

	less := func(a, b T) bool {
		return cmp(a, b) < 0
	}

	slices.SortFunc(newSlice, less)
	slices.SortFunc(oldSlice, less)

	i, j := 0, 0
	for i < len(oldSlice) && j < len(newSlice) {
		c := cmp(oldSlice[i], newSlice[j])
		switch {
		case c < 0: // removed
			removed = append(removed, oldSlice[i])
			i++

		case c > 0: // added
			added = append(added, newSlice[j])
			j++

		default: // kept
			kept = append(kept, newSlice[j])
			i++
			j++
		}
	}

	// Add rest

	for ; i < len(oldSlice); i++ {
		removed = append(removed, oldSlice[i])
	}

	for ; j < len(newSlice); j++ {
		added = append(added, newSlice[j])
	}

	return added, removed, kept
}

func Shuffle[T any](s []T) {
	rand.Shuffle(len(s), func(i, j int) {
		s[i], s[j] = s[j], s[i]
	})
}

func Filter[T any](s []T, cmp func(T) bool) []T {
	t := []T{}

	for _, i := range s {
		if cmp(i) {
			t = append(t, i)
		}
	}

	return t
}

func Contains[T any](s []T, cmp func(T) bool) bool {
	for _, i := range s {
		if cmp(i) {
			return true
		}
	}

	return false
}

func Map[T any](s []T, cb func(T) T) []T {
	n := []T{}

	for _, t := range s {
		n = append(n, cb(t))
	}

	return n
}

func String[T any](s []T) []string {
	n := []string{}

	for _, t := range s {
		n = append(n, fmt.Sprintf("%v", t))
	}

	return n
}
