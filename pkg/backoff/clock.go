// SPDX-FileCopyrightText: 2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package backoff

import "time"

type Clock interface {
	Sleep(duration time.Duration)
	Now() time.Time
}

type systemClock struct{}

func (c *systemClock) Sleep(d time.Duration) {
	time.Sleep(d)
}

func (c *systemClock) Now() time.Time {
	return time.Now()
}
