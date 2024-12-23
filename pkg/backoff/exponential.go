// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package backoff

import (
	"time"

	"cunicu.li/cunicu/pkg/crypto"
)

/*
ExponentialBackOff is a backoff implementation that increases the backoff
period for each retry attempt using a randomization function that grows exponentially.

NextBackOff() is calculated using the following formula:

	randomized interval =
	    RetryInterval * (random value in range [1 - RandomizationFactor, 1 + RandomizationFactor])

In other words NextBackOff() will range between the randomization factor
percentage below and above the retry interval.

For example, given the following parameters:

	RetryInterval = 2
	RandomizationFactor = 0.5
	Multiplier = 2

the actual backoff period used in the next retry attempt will range between 1 and 3 seconds,
multiplied by the exponential, that is, between 2 and 6 seconds.

Note: MaxInterval caps the RetryInterval and not the randomized interval.

If the time elapsed since an ExponentialBackOff instance is created goes past the
MaxElapsedTime, then the method NextBackOff() starts returning backoff.Stop.

The elapsed time can be reset by calling Reset().

Example: Given the following default arguments, for 10 tries the sequence will be,
and assuming we go over the MaxElapsedTime on the 10th try:

	Request #  RetryInterval (seconds)  Randomized Interval (seconds)

	 1          0.5                     [0.25,   0.75]
	 2          0.75                    [0.375,  1.125]
	 3          1.125                   [0.562,  1.687]
	 4          1.687                   [0.8435, 2.53]
	 5          2.53                    [1.265,  3.795]
	 6          3.795                   [1.897,  5.692]
	 7          5.692                   [2.846,  8.538]
	 8          8.538                   [4.269, 12.807]
	 9         12.807                   [6.403, 19.210]
	10         19.210                   backoff.Stop

Note: Implementation is not thread-safe.
*/
type ExponentialBackOff struct {
	InitialInterval     time.Duration
	RandomizationFactor float64
	Multiplier          float64
	MaxInterval         time.Duration
	// After MaxElapsedTime the ExponentialBackOff returns Stop.
	// It never stops if MaxElapsedTime == 0.
	MaxElapsedTime time.Duration

	CurrentInterval time.Duration
	StartTime       time.Time

	Clock Clock
}

// Reset the interval back to the initial retry interval and restarts the timer.
// Reset must be called before using b.
func (b *ExponentialBackOff) Reset() {
	if b.Clock == nil {
		b.Clock = &systemClock{}
	}

	b.CurrentInterval = b.InitialInterval
	b.StartTime = b.Clock.Now()
}

// NextBackOff calculates the next backoff interval using the formula:
//
//	Randomized interval = RetryInterval * (1 ± RandomizationFactor)
func (b *ExponentialBackOff) NextBackOff() time.Duration {
	// Make sure we have not gone over the maximum elapsed time.
	elapsed := b.GetElapsedTime()
	next := crypto.GetRandomValueFromInterval(b.RandomizationFactor, b.CurrentInterval)
	b.incrementCurrentInterval()

	if b.MaxElapsedTime != 0 && elapsed+next > b.MaxElapsedTime {
		return stop
	}

	return next
}

// GetElapsedTime returns the elapsed time since an ExponentialBackOff instance
// is created and is reset when Reset() is called.
//
// The elapsed time is computed using time.Now().UnixNano(). It is
// safe to call even while the backoff policy is used by a running
// ticker.
func (b *ExponentialBackOff) GetElapsedTime() time.Duration {
	return b.Clock.Now().Sub(b.StartTime)
}

// Increments the current interval by multiplying it with the multiplier.
func (b *ExponentialBackOff) incrementCurrentInterval() {
	// Check for overflow, if overflow is detected set the current interval to the max interval.
	if float64(b.CurrentInterval) >= float64(b.MaxInterval)/b.Multiplier {
		b.CurrentInterval = b.MaxInterval
	} else {
		b.CurrentInterval = time.Duration(float64(b.CurrentInterval) * b.Multiplier)
	}
}
