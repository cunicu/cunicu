package config

import (
	"errors"
)

type StructsProvider struct {
	value any
	tag   string
}

// StructsProvider is very similar koanf's struct provider
// but slightly adjusted to our needs.
func NewStructsProvider(v any, t string) *StructsProvider {
	return &StructsProvider{
		value: v,
		tag:   t,
	}
}

func (p *StructsProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *StructsProvider) Read() (map[string]any, error) {
	return Map(p.value, p.tag), nil
}
