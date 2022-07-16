package util_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/util"
)

var _ = Context("duration", func() {
	DescribeTable("test", func(dur time.Duration, output string) {
		Expect(util.PrettyDuration(dur, false)).To(Equal(output))
	},
		Entry("plural", 5*time.Hour+15*time.Minute+2*time.Second, "5 hours, 15 minutes, 2 seconds"),
		Entry("singular", time.Hour+time.Minute+time.Second, "1 hour, 1 minute, 1 second"),
		Entry("empty", 0*time.Second, ""),
	)
})

var _ = Specify("ago", func() {
	now := time.Now()

	Expect(util.Ago(now, false)).To(Equal("Now"))
	Expect(util.Ago(now.Add(-time.Hour), false)).To(Equal("1 hour ago"))
	Expect(util.Ago(now.Add(-time.Hour-10*time.Minute), false)).To(Equal("1 hour, 10 minutes ago"))
})

var _ = Context("bytes", func() {
	DescribeTable("test", func(bytes int, output string) {
		Expect(util.PrettyBytes(int64(bytes), false)).To(Equal(output))
	},
		Entry("without SI suffix", 500, "500 B"),
		Entry("without SI suffix", 1536, "1.50 KiB"),
		Entry("without SI suffix", 1572864, "1.50 MiB"),
	)
})

var _ = Context("every", func() {
	DescribeTable("test", func(dur time.Duration, output string) {
		Expect(util.Every(dur, false)).To(Equal(output))
	},
		Entry("plural", 5*time.Hour, "every 5 hours"),
		Entry("singular", time.Hour, "every 1 hour"),
	)
})
