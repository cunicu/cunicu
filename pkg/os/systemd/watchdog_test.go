// SPDX-FileCopyrightText: 2016 CoreOS, Inc.
// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package systemd_test

import (
	"os"
	"strconv"
	"time"

	"cunicu.li/cunicu/pkg/os/systemd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Watchdog", func() {
	myPID := strconv.Itoa(os.Getpid())

	DescribeTable("enabled", func(usec string, pid string, unsetEnv bool, expectedErr error, expectedDelay time.Duration) {
		if usec != "" {
			err := os.Setenv("WATCHDOG_USEC", usec)
			Expect(err).To(Succeed())
		} else {
			err := os.Unsetenv("WATCHDOG_USEC")
			Expect(err).To(Succeed())
		}

		if pid != "" {
			err := os.Setenv("WATCHDOG_PID", pid)
			Expect(err).To(Succeed())
		} else {
			err := os.Unsetenv("WATCHDOG_PID")
			Expect(err).To(Succeed())
		}

		delay, err := systemd.WatchdogEnabled(unsetEnv)
		Expect(delay).To(Equal(expectedDelay))

		if expectedErr != nil {
			Expect(err).To(MatchError(expectedErr))
		} else {
			Expect(err).To(Succeed())
		}

		if unsetEnv {
			Expect(os.Getenv("WATCHDOG_PID")).To(BeEmpty())
			Expect(os.Getenv("WATCHDOG_USEC")).To(BeEmpty())
		}
	},
		// Success cases
		Entry(nil, "100", myPID, true, nil, 100*time.Microsecond),
		Entry(nil, "50", myPID, true, nil, 50*time.Microsecond),
		Entry(nil, "1", myPID, false, nil, 1*time.Microsecond),
		Entry(nil, "1", "", true, nil, 1*time.Microsecond),

		// No-op cases
		Entry("WATCHDOG_USEC not set", "", myPID, true, nil, time.Duration(0)),
		Entry("WATCHDOG_PID doesn't match", "1", "0", false, nil, time.Duration(0)),
		Entry("Both not set", "", "", true, nil, time.Duration(0)),

		// Failure cases
		Entry("Negative USEC", "-1", myPID, true, systemd.ErrNegativeWatchdogInterval, time.Duration(0)),
		Entry("Non-integer USEC value", "string", "1", false, strconv.ErrSyntax, time.Duration(0)),
		Entry("Non-integer PID value", "1", "string", true, strconv.ErrSyntax, time.Duration(0)),
		Entry("Everything wrong", "stringa", "stringb", false, strconv.ErrSyntax, time.Duration(0)),
		Entry("Everything wrong", "-10239", "-eleventythree", true, systemd.ErrNegativeWatchdogInterval, time.Duration(0)),
	)
})
