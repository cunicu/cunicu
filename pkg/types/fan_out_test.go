// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"github.com/stv0g/cunicu/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("FanOut", func() {
	It("works with no channel", func() {
		fo := types.NewFanOut[int](1)

		fo.Send(1234)

		fo.Close()
	})

	It("works with a single channel", func() {
		fo := types.NewFanOut[int](1)
		ch := fo.Add()

		fo.Send(1234)

		Eventually(ch).Should(Receive(Equal(1234)))

		fo.Close()
	})

	It("works with two channels", func() {
		fo := types.NewFanOut[int](1)

		ch1 := fo.Add()
		ch2 := fo.Add()

		fo.Send(1234)

		Eventually(ch1).Should(Receive(Equal(1234)))
		Eventually(ch2).Should(Receive(Equal(1234)))

		fo.Close()
	})

	It("works with two channels after one has been removed", func() {
		fo := types.NewFanOut[int](1)
		ch1 := fo.Add()
		ch2 := fo.Add()

		fo.Remove(ch2)

		fo.Send(1234)

		Eventually(ch1).Should(Receive(Equal(1234)))

		fo.Close()
	})
})
