package config

import (
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
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
