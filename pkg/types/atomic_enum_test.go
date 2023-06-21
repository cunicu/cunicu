// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"github.com/stv0g/cunicu/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AtomicEnum", func() {
	type MyEnum int

	const (
		MyEnumBlue MyEnum = iota
		MyEnumRed
		MyEnumGreen
	)

	var e types.AtomicEnum[MyEnum]

	BeforeEach(func() {
		e = types.AtomicEnum[MyEnum]{}
	})

	It("is zero initialized", func() {
		Expect(e.Load()).To(Equal(MyEnumBlue))
	})

	It("can store", func() {
		e.Store(MyEnumGreen)
		Expect(e.Load()).To(Equal(MyEnumGreen))
	})

	It("can compare and swap 1", func() {
		swapped := e.CompareAndSwap(MyEnumGreen, MyEnumRed)
		Expect(swapped).To(BeFalse())
		Expect(e.Load()).To(Equal(MyEnumBlue))
	})

	It("can compare and swap 2", func() {
		swapped := e.CompareAndSwap(MyEnumBlue, MyEnumRed)
		Expect(swapped).To(BeTrue())
		Expect(e.Load()).To(Equal(MyEnumRed))
	})

	It("can swap", func() {
		e.Swap(MyEnumRed)
		Expect(e.Load()).To(Equal(MyEnumRed))
	})
})
