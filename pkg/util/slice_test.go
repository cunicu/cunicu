package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/util"
	"golang.org/x/exp/slices"
)

var _ = Context("slice", func() {
	var s []int

	BeforeEach(func() {
		s = []int{23, 54, 22, 8, 112, 234, 123}
	})

	It("map", func() {
		Expect(util.SliceMap(s, func(i int) int { return i + 100 })).To(Equal([]int{123, 154, 122, 108, 212, 334, 223}))
	})

	It("string", func() {
		Expect(util.SliceString(s)).To(Equal([]string{"23", "54", "22", "8", "112", "234", "123"}))
	})

	It("filter", func() {
		Expect(util.SliceFilter(s, func(i int) bool { return i != 22 && i != 123 })).To(Equal([]int{23, 54, 8, 112, 234}))
		Expect(util.SliceFilter(s, func(i int) bool { return true })).To(Equal(s))
	})

	It("contains", func() {
		Expect(util.SliceContains(s, func(i int) bool { return i == 8 })).To(BeTrue())
		Expect(util.SliceContains(s, func(i int) bool { return i == 9 })).To(BeFalse())
	})

	Describe("shuffle", func() {
		var a []int

		BeforeEach(func() {
			a = make([]int, 100)
			for i := 0; i < 100; i++ {
				a[i] = i
			}
		})

		It("shuffles properly", func() {
			Expect(slices.IsSorted(a)).To(BeTrue())
			Expect(a).To(HaveLen(100))

			util.SliceShuffle(a)

			Expect(slices.IsSorted(a)).NotTo(BeTrue())
			Expect(a).To(HaveLen(100))
		})
	})

	Context("diff", func() {
		var a, b, c, d, bc, cd []int
		var f func(old, new []int) (added, removed, kept []int)

		BeforeEach(func() {
			a = []int{}
			b = []int{1, 2, 3}
			c = []int{4, 5, 6}
			d = []int{7, 8, 9}

			bc = append(b, c...)
			cd = append(c, d...)

			util.SliceShuffle(b)
			util.SliceShuffle(c)
			util.SliceShuffle(d)
			util.SliceShuffle(bc)
			util.SliceShuffle(cd)
		})

		Test := func() {
			It("finds added elements", func() {
				added, removed, kept := f(a, b)

				Expect(added).To(ConsistOf(b))
				Expect(removed).To(BeEmpty())
				Expect(kept).To(BeEmpty())
			})

			It("finds removed elements", func() {
				added, removed, kept := f(b, a)

				Expect(added).To(BeEmpty())
				Expect(removed).To(ConsistOf(b))
				Expect(kept).To(BeEmpty())
			})

			It("finds kept elements", func() {
				added, removed, kept := f(a, a)

				Expect(added).To(BeEmpty())
				Expect(removed).To(BeEmpty())
				Expect(kept).To(ConsistOf(a))
			})

			It("finds no changes on empty slices", func() {
				added, removed, kept := f(a, a)

				Expect(added).To(BeEmpty())
				Expect(removed).To(BeEmpty())
				Expect(kept).To(BeEmpty())
			})

			It("finds added and removed elements", func() {
				added, removed, kept := f(b, c)

				Expect(added).To(ConsistOf(c))
				Expect(removed).To(ConsistOf(b))
				Expect(kept).To(BeEmpty())
			})

			It("finds all changes at once", func() {
				added, removed, kept := f(bc, cd)

				Expect(added).To(ConsistOf(d))
				Expect(removed).To(ConsistOf(b))
				Expect(kept).To(ConsistOf(c))
			})
		}

		Describe("func", func() {
			BeforeEach(func() {
				f = func(old, new []int) (added, removed, kept []int) {
					return util.SliceDiffFunc(old, new, func(a, b int) int {
						return a - b
					})
				}
			})

			Test()
		})

		Describe("no func", func() {
			BeforeEach(func() {
				f = util.SliceDiff[int]
			})

			Test()
		})
	})
})
