// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

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
	return nil, errNotImplemented
}

func (p *StructsProvider) Read() (map[string]any, error) {
	return Map(p.value, p.tag), nil
}
