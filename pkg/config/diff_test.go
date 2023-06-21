// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"github.com/stv0g/cunicu/pkg/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("diff", func() {
	It("finds no changes in equal configs", func() {
		c := &config.Settings{
			RPC: config.RPCSettings{
				Socket: "/path/cunicu.sock",
			},
		}

		changes := config.DiffSettings(c, c)
		Expect(changes).To(HaveLen(0))
	})

	// It("can detect added settings", func() {
	// 	old := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 	})

	// 	new := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 		"setting3.subsetting2": 3,
	// 	})

	// 	changes := config.DiffSettings(old, new)
	// 	Expect(changes).To(HaveKeyWithValue("setting3.subsetting2", config.Change{
	// 		Old: nil,
	// 		New: 3,
	// 	}))
	// })

	// It("can detect added setting group", func() {
	// 	old := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 	})

	// 	new := createConfig(map[string]any{
	// 		"setting1":                           1,
	// 		"setting2":                           "hallo",
	// 		"setting3.subsetting1":               2,
	// 		"setting3.subsettinggroup2.setting1": 3,
	// 	})

	// 	changes := config.DiffSettings(old, new)
	// 	Expect(changes).To(HaveKeyWithValue("setting3.subsettinggroup2", config.Change{
	// 		Old: nil,
	// 		New: map[string]any{"setting1": 3},
	// 	}))
	// })

	// It("can detect removed setting group", func() {
	// 	old := createConfig(map[string]any{
	// 		"setting1":                           1,
	// 		"setting2":                           "hallo",
	// 		"setting3.subsetting1":               2,
	// 		"setting3.subsettinggroup2.setting1": 3,
	// 	})

	// 	new := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 	})

	// 	changes := config.DiffSettings(old, new)
	// 	Expect(changes).To(HaveKeyWithValue("setting3.subsettinggroup2", config.Change{
	// 		Old: map[string]any{"setting1": 3},
	// 		New: nil,
	// 	}))
	// })

	// It("can detect removed settings", func() {
	// 	old := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 		"setting3.subsetting2": 3,
	// 	})

	// 	new := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 	})

	// 	changes := config.DiffSettings(old, new)
	// 	Expect(changes).To(HaveKeyWithValue("setting3.subsetting2", config.Change{
	// 		Old: 3,
	// 		New: nil,
	// 	}))
	// })

	// It("can detect changed settings", func() {
	// 	old := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 		"setting3.subsetting2": 3,
	// 	})

	// 	new := createConfig(map[string]any{
	// 		"setting1":             1,
	// 		"setting2":             "hallo",
	// 		"setting3.subsetting1": 2,
	// 		"setting3.subsetting2": 4,
	// 	})

	// 	changes := config.DiffSettings(old, new)
	// 	Expect(changes).To(HaveKeyWithValue("setting3.subsetting2", config.Change{
	// 		Old: 3,
	// 		New: 4,
	// 	}))
	// })

	// It("can detect changed setting group", func() {
	// 	old := createConfig(map[string]any{
	// 		"setting1":                  1,
	// 		"setting2":                  "hallo",
	// 		"setting3.subsetting1":      2,
	// 		"setting3.subsettinggroup2": 3,
	// 	})

	// 	new := createConfig(map[string]any{
	// 		"setting1":                           1,
	// 		"setting2":                           "hallo",
	// 		"setting3.subsetting1":               2,
	// 		"setting3.subsettinggroup2.setting1": 4,
	// 		"setting3.subsettinggroup2.setting2": 5,
	// 	})

	// 	changes := config.DiffSettings(old, new)
	// 	Expect(changes).To(HaveKeyWithValue("setting3.subsettinggroup2", config.Change{
	// 		Old: 3,
	// 		New: map[string]any{
	// 			"setting1": 4,
	// 			"setting2": 5,
	// 		},
	// 	}))
	// })
})
