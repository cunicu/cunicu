package util_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/pkg/util"
)

var _ = Context("human", func() {
	Context("duration", func() {
		DescribeTable("test", func(dur time.Duration, output string) {
			Expect(util.StripANSI(util.PrettyDuration(dur))).To(Equal(output))
		},
			Entry("plural", 5*time.Hour+15*time.Minute+2*time.Second, "5 hours, 15 minutes, 2 seconds"),
			Entry("singular", time.Hour+time.Minute+time.Second, "1 hour, 1 minute, 1 second"),
			Entry("empty", 0*time.Second, ""),
		)
	})

	Specify("ago", func() {
		now := time.Now()

		Expect(util.StripANSI(util.Ago(now))).To(Equal("Now"))
		Expect(util.StripANSI(util.Ago(now.Add(-time.Hour)))).To(Equal("1 hour ago"))
		Expect(util.StripANSI(util.Ago(now.Add(-time.Hour - 10*time.Minute)))).To(Equal("1 hour, 10 minutes ago"))
	})

	Context("bytes", func() {
		DescribeTable("test", func(bytes int, output string) {
			Expect(util.StripANSI(util.PrettyBytes(int64(bytes)))).To(Equal(output))
		},
			Entry("without SI suffix", 500, "500 B"),
			Entry("without SI suffix", 1536, "1.50 KiB"),
			Entry("without SI suffix", 1572864, "1.50 MiB"),
		)
	})

	Context("every", func() {
		DescribeTable("test", func(dur time.Duration, output string) {
			Expect(util.StripANSI(util.Every(dur))).To(Equal(output))
		},
			Entry("plural", 5*time.Hour, "every 5 hours"),
			Entry("singular", time.Hour, "every 1 hour"),
		)
	})
})
