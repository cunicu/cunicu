// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package os

import (
	"time"
)

// GetClockMonotonic returns the current time from the CLOCK_MONOTONIC clock.
func GetClockMonotonic() (t time.Time, err error) {
	return time.Now(), nil
}
