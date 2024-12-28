// SPDX-FileCopyrightText: 2015 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package systemd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

var ErrNegativeWatchdogInterval = errors.New("WATCHDOG_USEC must be a positive number")

// WatchdogEnabled returns watchdog information for a service.
// Processes should call daemon.SdNotify(false, daemon.SdNotifyWatchdog) every
// time / 2.
// If `unsetEnv` is true, the environment variables `WATCHDOG_USEC` and
// `WATCHDOG_PID` will be unconditionally unset.
//
// It returns one of the following:
// (0, nil) - watchdog isn't enabled or we aren't the watched PID.
// (0, err) - an error happened (e.g. error converting time).
// (time, nil) - watchdog is enabled and we can send ping.  time is delay
// before inactive service will be killed.
func WatchdogEnabled(unsetEnv bool) (interval time.Duration, err error) {
	wUSec := os.Getenv("WATCHDOG_USEC")
	wPID := os.Getenv("WATCHDOG_PID")

	if unsetEnv {
		wUSecErr := os.Unsetenv("WATCHDOG_USEC")
		wPIDErr := os.Unsetenv("WATCHDOG_PID")

		if wUSecErr != nil {
			return 0, wUSecErr
		}

		if wPIDErr != nil {
			return 0, wPIDErr
		}
	}

	if wUSec == "" {
		return 0, nil
	}

	s, err := strconv.Atoi(wUSec)
	if err != nil {
		return 0, fmt.Errorf("failed to convert WATCHDOG_USEC: %w", err)
	} else if s <= 0 {
		return 0, ErrNegativeWatchdogInterval
	}

	interval = time.Duration(s) * time.Microsecond

	if wPID == "" {
		return interval, nil
	}

	if p, err := strconv.Atoi(wPID); err != nil {
		return 0, fmt.Errorf("failed to convert WATCHDOG_PID: %w", err)
	} else if os.Getpid() != p {
		return 0, nil
	}

	return interval, nil
}
