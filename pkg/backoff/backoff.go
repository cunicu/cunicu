// SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package backoff

import (
	"iter"
	"time"
)

// stop indicates that no more retries should be made for use in NextBackOff().
const stop time.Duration = -1

func Retry(b *ExponentialBackOff) iter.Seq2[int, time.Duration] {
	return func(yield func(int, time.Duration) bool) {
		b.Reset()

		for i := 0; ; i++ {
			if !yield(i, b.GetElapsedTime()) {
				break
			}

			next := b.NextBackOff()
			if next == stop {
				break
			}

			b.Clock.Sleep(next)
		}
	}
}
