// SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package backoff_test

import (
	"testing"
	"time"

	"cunicu.li/cunicu/pkg/backoff"
	"cunicu.li/cunicu/test"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockClock struct {
	now time.Time
}

func (c *mockClock) Sleep(d time.Duration) {
	c.now = c.now.Add(d)
}

func (c *mockClock) Now() time.Time {
	return c.now
}

func TestSuite(t *testing.T) {
	test.SetupLogging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Backoff Suite")
}

var _ = Context("Backoff", func() {
	var (
		b = &backoff.ExponentialBackOff{
			InitialInterval:     100 * time.Millisecond,
			RandomizationFactor: 0.1,
			Multiplier:          3,
			MaxInterval:         10 * time.Second,
			MaxElapsedTime:      35 * time.Second,

			Clock: &mockClock{},
		}

		expectedResults = []time.Duration{
			100 * time.Millisecond,
			300 * time.Millisecond,
			900 * time.Millisecond,
			2700 * time.Millisecond,
			8100 * time.Millisecond,
			10000 * time.Millisecond,
			10000 * time.Millisecond,
			10000 * time.Millisecond,
			10000 * time.Millisecond,
		}
	)

	It("produces exponentially increasing backoff durations", func() {
		b.Reset()

		for _, expected := range expectedResults {
			// Assert that the next backoff falls in the expected range.
			Expect(b.CurrentInterval).To(BeNumerically("==", expected))

			minInterval := expected - time.Duration(b.RandomizationFactor*float64(expected))
			maxInterval := expected + time.Duration(b.RandomizationFactor*float64(expected))

			actualInterval := b.NextBackOff()
			Expect(actualInterval).To(BeNumerically(">=", minInterval))
			Expect(actualInterval).To(BeNumerically("<=", maxInterval))
		}
	})

	Context("Go function interator", func() {
		It("stops", func() {
			var i int
			for i = range backoff.Retry(b) {
				expected := expectedResults[i]

				Expect(b.CurrentInterval).To(BeNumerically("==", expected))
			}

			Expect(i).To(Equal(7))
		})

		It("can break out early", func() {
			j := 0
			for i := range backoff.Retry(b) {
				expected := expectedResults[i]

				Expect(b.CurrentInterval).To(BeNumerically("==", expected))

				if j >= 6 {
					break
				}

				j++
			}

			Expect(j).To(Equal(6))
		})
	})
})
