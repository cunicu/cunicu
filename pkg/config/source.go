// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

type Source interface {
	Load() error
	Config() *koanf.Koanf
	Order() []string
}

type source struct {
	*koanf.Koanf
	koanf.Provider

	lastVersion any
}

func (s *source) Order() []string {
	order := []string{}

	if q, ok := s.Provider.(Orderable); ok {
		order = append(order, q.Order()...)
	}

	if s, ok := s.Provider.(SubProvidable); ok {
		for _, p := range s.SubProviders() {
			if q, ok := p.(Orderable); ok {
				order = append(order, q.Order()...)
			}
		}
	}

	return order
}

func (s *source) Config() *koanf.Koanf {
	return s.Koanf
}

func (s *source) Load() error {
	var (
		err     error
		version any
	)

	if v, ok := s.Provider.(Versioned); ok {
		version = v.Version()

		// Do not reload if we already are loaded and the version has not changed
		if s.Koanf != nil && version == s.lastVersion {
			return nil
		}
	}

	if s.Koanf, err = load(s.Provider); err != nil {
		return err
	}

	s.lastVersion = version

	return nil
}

func load(p koanf.Provider) (*koanf.Koanf, error) {
	var q koanf.Parser
	switch p.(type) {
	case *RemoteFileProvider, *LocalFileProvider, *rawbytes.RawBytes:
		q = yaml.Parser()
	default:
		q = nil
	}

	k := koanf.New(".")

	if err := k.Load(p, q); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if s, ok := p.(SubProvidable); ok {
		for _, p := range s.SubProviders() {
			d, err := load(p)
			if err != nil {
				return nil, err
			}

			if err := k.Merge(d); err != nil {
				return nil, fmt.Errorf("failed to merge config: %w", err)
			}
		}
	}

	return k, nil
}
