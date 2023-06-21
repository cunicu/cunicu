// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package tty_test

import (
	"time"

	"github.com/stv0g/cunicu/pkg/tty"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("human", func() {
	Context("duration", func() {
		DescribeTable("test",
			func(dur time.Duration, output string) {
				Expect(tty.StripANSI(tty.PrettyDuration(dur))).To(Equal(output))
			},
			Entry("plural", 5*time.Hour+15*time.Minute+2*time.Second, "5 hours, 15 minutes, 2 seconds"),
			Entry("singular", time.Hour+time.Minute+time.Second, "1 hour, 1 minute, 1 second"),
			Entry("empty", 0*time.Second, ""),
		)
	})

	Specify("ago", func() {
		now := time.Now()

		Expect(tty.StripANSI(tty.Ago(now))).To(Equal("Now"))
		Expect(tty.StripANSI(tty.Ago(now.Add(-time.Hour)))).To(Equal("1 hour ago"))
		Expect(tty.StripANSI(tty.Ago(now.Add(-time.Hour - 10*time.Minute)))).To(Equal("1 hour, 10 minutes ago"))
	})

	Context("bytes", func() {
		DescribeTable("test",
			func(bytes int, output string) {
				Expect(tty.StripANSI(tty.PrettyBytes(int64(bytes)))).To(Equal(output))
			},
			Entry("on boundary", 1024, "1.00 KiB"),
			Entry("without SI suffix", 500, "500 B"),
			Entry("without SI suffix", 1536, "1.50 KiB"),
			Entry("without SI suffix", 1572864, "1.50 MiB"),
		)
	})

	Context("every", func() {
		DescribeTable("test",
			func(dur time.Duration, output string) {
				Expect(tty.StripANSI(tty.Every(dur))).To(Equal(output))
			},
			Entry("plural", 5*time.Hour, "every 5 hours"),
			Entry("singular", time.Hour, "every 1 hour"),
		)
	})
})
