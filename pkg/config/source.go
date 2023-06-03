// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/v2"
)

type Source struct {
	Provider koanf.Provider
	Config   *koanf.Koanf

	Order []string

	LastVersion any
}

func (s *Source) Load() error {
	var err error
	var version any
	if v, ok := s.Provider.(Versioned); ok {
		version = v.Version()

		// Do not reload if we already are loaded and the version has not changed
		if s.Config != nil && version == s.LastVersion {
			return nil
		}
	}

	if s.Config, s.Order, err = load(s.Provider); err != nil {
		return err
	}

	s.LastVersion = version

	return nil
}

func load(p koanf.Provider) (*koanf.Koanf, []string, error) {
	var q koanf.Parser
	switch p.(type) {
	case *RemoteFileProvider, *LocalFileProvider:
		q = yaml.Parser()
	default:
		q = nil
	}

	k := koanf.New(".")
	o := []string{}

	if err := k.Load(p, q); err != nil {
		return nil, nil, fmt.Errorf("failed to load config: %w", err)
	}

	if q, ok := p.(Orderable); ok {
		o = append(o, q.Order()...)
	}

	if w, ok := p.(Watchable); ok {
		if err := w.Watch(func(event interface{}, err error) {}); err != nil {
			if !strings.Contains(err.Error(), "does not support this method") {
				return nil, nil, fmt.Errorf("failed to watch for changes: %w", err)
			}
		}
	}

	if s, ok := p.(SubProvidable); ok {
		for _, p := range s.SubProviders() {
			d, m, err := load(p)
			if err != nil {
				return nil, nil, err
			}

			if err := k.Merge(d); err != nil {
				return nil, nil, fmt.Errorf("failed to merge config: %w", err)
			}

			o = append(o, m...)
		}
	}

	return k, o, nil
}
