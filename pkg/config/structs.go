package config

import "errors"

type structsProvider struct {
	value any
	tag   string
}

// StructsProvider is very similar koanf's struct provider
// but slightly adjusted to our needs.
func StructsProvider(v any, t string) *structsProvider {
	return &structsProvider{
		value: v,
		tag:   t,
	}
}

func (p *structsProvider) ReadBytes() ([]byte, error) {
	return nil, errors.New("this provider requires no parser")
}

func (p *structsProvider) Read() (map[string]any, error) {
	return Map(p.value, p.tag), nil
}
