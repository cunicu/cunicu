// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package types_test

import (
	"github.com/stv0g/cunicu/pkg/types"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("DiffMapp", func() {
	It("finds no changes in equal maps", func() {
		c := map[string]any{
			"a": "b",
		}

		changes := types.DiffMap(c, c)
		Expect(changes).To(HaveLen(0))
	})

	It("can detect added keys", func() {
		oldMap := map[string]any{
			"a": 1,
		}

		newMap := map[string]any{
			"a": 1,
			"b": 2,
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"b": {
				New: 2,
			},
		}))
	})

	It("can detect removed keys", func() {
		oldMap := map[string]any{
			"a": 1,
			"b": 2,
		}

		newMap := map[string]any{
			"a": 1,
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"b": {
				Old: 2,
			},
		}))
	})

	It("can detect changed keys", func() {
		oldMap := map[string]any{
			"a": 1,
		}

		newMap := map[string]any{
			"a": 2,
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a": {
				Old: 1,
				New: 2,
			},
		}))
	})

	It("can detect changed keys with slice values", func() {
		oldMap := map[string]any{
			"a": []string{"c"},
		}

		newMap := map[string]any{
			"a": []string{"c", "d"},
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a": {
				Old: []string{"c"},
				New: []string{"c", "d"},
			},
		}))
	})

	It("can detect added and removed keys", func() {
		oldMap := map[string]any{
			"a": 1,
		}

		newMap := map[string]any{
			"b": 1,
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a": {
				Old: 1,
			},
			"b": {
				New: 1,
			},
		}))
	})

	It("can detect added keys in group", func() {
		oldMap := map[string]any{}

		newMap := map[string]any{
			"a.b": 1,
			"b": map[string]any{
				"c": 2,
			},
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a": {
				New: map[string]any{
					"b": 1,
				},
			},
			"b": {
				New: map[string]any{
					"c": 2,
				},
			},
		}))
	})

	It("can detect removed keys in group", func() {
		oldMap := map[string]any{
			"a.b": 1,
			"b": map[string]any{
				"c": 2,
			},
		}

		newMap := map[string]any{}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a": {
				Old: map[string]any{
					"b": 1,
				},
			},
			"b": {
				Old: map[string]any{
					"c": 2,
				},
			},
		}))
	})

	It("can detect changed keys in group", func() {
		oldMap := map[string]any{
			"a.b": 1,
			"b": map[string]any{
				"c": 2,
			},
		}

		newMap := map[string]any{
			"a.b": 3,
			"b": map[string]any{
				"c": 4,
			},
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a.b": {
				Old: 1,
				New: 3,
			},
			"b.c": {
				Old: 2,
				New: 4,
			},
		}))
	})

	It("can detect added group", func() {
		oldMap := map[string]any{
			"a.b": 1,
		}

		newMap := map[string]any{
			"a.b": 1,
			"b": map[string]any{
				"c": 2,
				"d": 3,
			},
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"b": {
				New: map[string]any{
					"c": 2,
					"d": 3,
				},
			},
		}))
	})

	It("can detect removed group", func() {
		oldMap := map[string]any{
			"a.b": 1,
			"b": map[string]any{
				"c": 2,
				"d": 3,
			},
		}

		newMap := map[string]any{
			"a.b": 1,
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"b": {
				Old: map[string]any{
					"c": 2,
					"d": 3,
				},
			},
		}))
	})

	It("can detect changed group", func() {
		oldMap := map[string]any{
			"a.b": 1,
			"a.c": 2,
		}

		newMap := map[string]any{
			"a": map[string]any{
				"b": 4,
				"d": 5,
			},
		}

		changes := types.DiffMap(oldMap, newMap)
		Expect(changes).To(Equal(map[string]types.Change{
			"a.b": {
				Old: 1,
				New: 4,
			},
			"a.c": {
				Old: 2,
			},
			"a.d": {
				New: 5,
			},
		}))
	})
})
