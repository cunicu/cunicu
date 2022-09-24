package config

import "errors"

type confmapProvider struct {
	value any
}

// ConfMapProvider is very similar koanf's confmap provider
// but slightly adjusted to our needs.
func ConfMapProvider(v any) *confmapProvider {
	return &confmapProvider{
		value: v,
	}
}

func (p *confmapProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *confmapProvider) Read() (map[string]any, error) {
	return Map(p.value, "koanf"), nil
}
