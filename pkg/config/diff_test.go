package config_test

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/confmap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stv0g/cunicu/pkg/config"
)

var _ = Context("diff", func() {
	createConfig := func(m map[string]any) *koanf.Koanf {
		c := koanf.New(".")

		err := c.Load(confmap.Provider(m, "."), nil)
		Expect(err).To(Succeed())

		return c
	}

	It("finds no changes in equal configs", func() {
		c := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
			"setting3.subsetting2": 3,
		})

		changes := config.DiffConfig(c, c)
		Expect(changes).To(HaveLen(0))
	})

	It("can detect added settings", func() {
		old := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
		})

		new := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
			"setting3.subsetting2": 3,
		})

		changes := config.DiffConfig(old, new)
		Expect(changes).To(HaveKeyWithValue("setting3.subsetting2", config.Change{
			Old: nil,
			New: 3,
		}))
	})

	It("can detect added setting group", func() {
		old := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
		})

		new := createConfig(map[string]any{
			"setting1":                           1,
			"setting2":                           "hallo",
			"setting3.subsetting1":               2,
			"setting3.subsettinggroup2.setting1": 3,
		})

		changes := config.DiffConfig(old, new)
		Expect(changes).To(HaveKeyWithValue("setting3.subsettinggroup2", config.Change{
			Old: nil,
			New: map[string]any{"setting1": 3},
		}))
	})

	It("can detect removed setting group", func() {
		old := createConfig(map[string]any{
			"setting1":                           1,
			"setting2":                           "hallo",
			"setting3.subsetting1":               2,
			"setting3.subsettinggroup2.setting1": 3,
		})

		new := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
		})

		changes := config.DiffConfig(old, new)
		spew.Dump(changes)
		Expect(changes).To(HaveKeyWithValue("setting3.subsettinggroup2", config.Change{
			Old: map[string]any{"setting1": 3},
			New: nil,
		}))
	})

	It("can detect removed settings", func() {
		old := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
			"setting3.subsetting2": 3,
		})

		new := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
		})

		changes := config.DiffConfig(old, new)
		Expect(changes).To(HaveKeyWithValue("setting3.subsetting2", config.Change{
			Old: 3,
			New: nil,
		}))
	})

	It("can detect changed settings", func() {
		old := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
			"setting3.subsetting2": 3,
		})

		new := createConfig(map[string]any{
			"setting1":             1,
			"setting2":             "hallo",
			"setting3.subsetting1": 2,
			"setting3.subsetting2": 4,
		})

		changes := config.DiffConfig(old, new)
		Expect(changes).To(HaveKeyWithValue("setting3.subsetting2", config.Change{
			Old: 3,
			New: 4,
		}))
	})

	It("can detect changed setting group", func() {
		old := createConfig(map[string]any{
			"setting1":                  1,
			"setting2":                  "hallo",
			"setting3.subsetting1":      2,
			"setting3.subsettinggroup2": 3,
		})

		new := createConfig(map[string]any{
			"setting1":                           1,
			"setting2":                           "hallo",
			"setting3.subsetting1":               2,
			"setting3.subsettinggroup2.setting1": 4,
			"setting3.subsettinggroup2.setting2": 5,
		})

		changes := config.DiffConfig(old, new)
		Expect(changes).To(HaveKeyWithValue("setting3.subsettinggroup2", config.Change{
			Old: 3,
			New: map[string]any{
				"setting1": 4,
				"setting2": 5,
			},
		}))
	})
})
