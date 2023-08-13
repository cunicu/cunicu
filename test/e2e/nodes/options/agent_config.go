// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package options

import (
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"

	"cunicu.li/cunicu/pkg/config"
)

type Config func(k *koanf.Koanf)

func (c Config) Apply(k *koanf.Koanf) {
	c(k)
}

type ConfigMap map[string]any

func (m ConfigMap) Apply(k *koanf.Koanf) {
	p := confmap.Provider(m, ".")

	k.Load(p, nil) //nolint:errcheck
}

func ConfigValue(key string, value any) Config {
	return func(k *koanf.Koanf) {
		k.Set(key, value) //nolint:errcheck
	}
}

func ConfigStruct(s *config.Settings) Config {
	return func(k *koanf.Koanf) {
		p := config.NewStructsProvider(s, "koanf")

		k.Load(p, nil) //nolint:errcheck
	}
}
