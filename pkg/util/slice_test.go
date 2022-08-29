package util_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/exp/slices"
	"riasc.eu/wice/pkg/util"
)

var _ = Context("Slice", func() {
	Describe("Shuffle", func() {
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

			util.ShuffleSlice(a)

			Expect(slices.IsSorted(a)).NotTo(BeTrue())
			Expect(a).To(HaveLen(100))
		})
	})

	Describe("Diff", func() {
		cmp := func(a, b *int) int {
			return *a - *b
		}

		var a, b, c, d, bc, cd []int

		BeforeEach(func() {
			a = []int{}
			b = []int{1, 2, 3}
			c = []int{4, 5, 6}
			d = []int{7, 8, 9}

			bc = append(b, c...)
			cd = append(c, d...)

			util.ShuffleSlice(b)
			util.ShuffleSlice(c)
			util.ShuffleSlice(d)
			util.ShuffleSlice(bc)
			util.ShuffleSlice(cd)
		})

		It("finds added elements", func() {
			added, removed, kept := util.DiffSliceFunc(a, b, cmp)

			Expect(added).To(ConsistOf(b))
			Expect(removed).To(BeEmpty())
			Expect(kept).To(BeEmpty())
		})

		It("finds removed elements", func() {
			added, removed, kept := util.DiffSliceFunc(b, a, cmp)

			Expect(added).To(BeEmpty())
			Expect(removed).To(ConsistOf(b))
			Expect(kept).To(BeEmpty())
		})

		It("finds kept elements", func() {
			added, removed, kept := util.DiffSliceFunc(a, a, cmp)

			Expect(added).To(BeEmpty())
			Expect(removed).To(BeEmpty())
			Expect(kept).To(ConsistOf(a))
		})

		It("finds no changes on empty slices", func() {
			added, removed, kept := util.DiffSliceFunc(a, a, cmp)

			Expect(added).To(BeEmpty())
			Expect(removed).To(BeEmpty())
			Expect(kept).To(BeEmpty())
		})

		It("finds added and removed elements", func() {
			added, removed, kept := util.DiffSliceFunc(b, c, cmp)

			Expect(added).To(ConsistOf(c))
			Expect(removed).To(ConsistOf(b))
			Expect(kept).To(BeEmpty())
		})

		It("finds all changes at once", func() {
			added, removed, kept := util.DiffSliceFunc(bc, cd, cmp)

			Expect(added).To(ConsistOf(d))
			Expect(removed).To(ConsistOf(b))
			Expect(kept).To(ConsistOf(c))
		})
	})
})
