// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const (
	envPrefix = "CUNICU_"
)

var errUnsupportedScheme = errors.New("unsupported scheme")

type Watchable interface {
	Watch(cb func(event interface{}, err error)) error
}

type Orderable interface {
	Order() []string
}

type SubProvidable interface {
	SubProviders() []koanf.Provider
}

type Versioned interface {
	Version() any
}

type Provider struct {
	koanf.Provider

	Config *koanf.Koanf
}

// Load loads configuration settings from various sources
//
// Settings are loaded in the following order where the later overwrite the previous settings:
// - defaults
// - dns lookups
// - configuration files
// - environment variables
// - command line flags
func (c *Config) GetProviders() ([]koanf.Provider, error) {
	ps := []koanf.Provider{
		NewStructsProvider(&DefaultSettings, "koanf"),
		NewWireGuardProvider(),
	}

	// Load settings from DNS lookups
	for _, domain := range c.Domains {
		p := NewLookupProvider(domain)
		ps = append(ps, p)
	}

	// Search for config files
	if len(c.Files) == 0 {
		searchPath := []string{"/etc", "/etc/cunicu"}

		if cwd, err := os.Getwd(); err != nil {
			c.logger.Warn("Failed to get working directory", zap.Error(err))
		} else {
			searchPath = append(searchPath, cwd)
		}

		if cfgDir := os.Getenv("CUNICU_CONFIG_DIR"); cfgDir != "" {
			searchPath = append(searchPath, cfgDir)
		}

		for _, path := range searchPath {
			fn := filepath.Join(path, "cunicu.yaml")
			if fi, err := os.Stat(fn); err == nil && !fi.IsDir() {
				c.Files = append(c.Files, fn)
			}
		}
	}

	// Add config files providers
	for _, f := range c.Files {
		u, err := url.Parse(f)
		if err != nil {
			return nil, fmt.Errorf("ignoring config file with invalid name: %w", err)
		}

		var p koanf.Provider
		switch u.Scheme {
		case "http", "https":
			p = NewRemoteFileProvider(u)
		case "":
			p = NewLocalFileProvider(u.Path)
		default:
			if isWindowsDriveLetter(u.Scheme) {
				p = NewLocalFileProvider(f)
			} else {
				return nil, fmt.Errorf("%w '%s' for config file", errUnsupportedScheme, u.Scheme)
			}
		}

		ps = append(ps, p)
	}

	// Add a runtime configuration file if it exists
	if fi, err := os.Stat(RuntimeConfigFile); err == nil && !fi.IsDir() {
		ps = append(ps,
			NewLocalFileProvider(RuntimeConfigFile),
		)
	}

	ps = append(ps,
		c.EnvironmentProvider(),
		c.FlagProvider(),
	)

	return ps, nil
}
