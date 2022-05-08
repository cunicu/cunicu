package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"riasc.eu/wice/internal/util"
)

var _ = Specify("Fan-in", func() {
	N := 5

	ch_in := []chan int{}
	for i := 0; i < N; i++ {
		ch_in = append(ch_in, make(chan int))
	}

	ch_out := util.FanIn(ch_in...)
	Expect(ch_out).NotTo(BeNil())
	Expect(ch_out).NotTo(BeClosed())

	for i := 0; i < N; i++ {
		ch_in[i] <- i
	}

	for i := 0; i < N; i++ {
		Eventually(ch_out).Should(Receive(Equal(i)))
	}
})
