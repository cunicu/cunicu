// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

func (c *Config) EnvironmentProvider() koanf.Provider {
	// Load environment variables
	envKeyMap := map[string]string{}
	for _, k := range c.Meta.Keys() {
		m := strings.ToUpper(k)
		e := envPrefix + strings.ReplaceAll(m, ".", "_")
		envKeyMap[e] = k
	}

	return env.ProviderWithValue(envPrefix, ".", func(e, v string) (string, any) {
		k := envKeyMap[e]

		if p := strings.Split(v, ","); len(p) > 1 {
			return k, p
		}

		return k, v
	})
}
