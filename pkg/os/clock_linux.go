// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package os

import (
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// GetClockMonotonic returns the current time from the CLOCK_MONOTONIC clock.
func GetClockMonotonic() (t time.Time, err error) {
	var ts syscall.Timespec
	if _, _, err := syscall.Syscall(syscall.SYS_CLOCK_GETTIME, unix.CLOCK_MONOTONIC, uintptr(unsafe.Pointer(&ts)), 0); err != 0 {
		return time.Time{}, err
	}

	return time.Unix(int64(ts.Sec), int64(ts.Nsec)), nil //nolint:unconvert
}
