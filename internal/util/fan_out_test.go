package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/util"
)

var _ = Describe("Fan-out", func() {
	It("should require buffered channels if we synchronously receive from the output channels", func() {
		fo := util.NewFanout[int](1)

		ch1 := fo.Add()
		ch2 := fo.Add()

		fo.C <- 1234

		Eventually(ch1).Should(Receive(Equal(1234)))
		Eventually(ch2).Should(Receive(Equal(1234)))
	})

	It("also works with unbuffered channels if there is only a single channel", func() {
		fo := util.NewFanout[int](0)
		ch := fo.Add()

		fo.C <- 1234

		Eventually(ch).Should(Receive(Equal(1234)))
	})

	It("might deadlock if there are more reciving channels", func() {
		fo := util.NewFanout[int](0)
		ch1 := fo.Add()
		ch2 := fo.Add()

		fo.C <- 1234

		Eventually(ch1).ShouldNot(Receive(Equal(1234)))
		Eventually(ch2).ShouldNot(Receive(Equal(1234)))
	})
})
