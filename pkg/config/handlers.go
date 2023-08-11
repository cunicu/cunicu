// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strings"

	"golang.org/x/exp/slices"

	"github.com/stv0g/cunicu/pkg/types"
)

type ChangedHandler interface {
	OnConfigChanged(key string, oldValue, newValue any) error
}

func (c *Config) InvokeChangedHandlers(key string, change types.Change) error {
	if err := c.Meta.InvokeChangedHandlers(key, change); err != nil {
		return err
	}

	// Invoke handlers for per-interface settings
	if keyParts := strings.Split(key, "."); len(keyParts) > 0 && keyParts[0] == "interfaces" {
		pattern := keyParts[1]

		for name, meta := range c.onInterfaceChanged {
			pats := c.InterfaceOrderByName(name)

			if slices.Contains(pats, pattern) {
				key := strings.Join(keyParts[2:], ".")

				if err := meta.InvokeChangedHandlers(key, change); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
