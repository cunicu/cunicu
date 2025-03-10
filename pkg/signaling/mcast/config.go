// SPDX-FileCopyrightText: 2025 Adam Rizkalla <ajarizzo@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package mcast

import (
	"fmt"
	"net"

	"cunicu.li/cunicu/pkg/signaling"
)

type BackendOptions struct {
	Interface *net.Interface
	Loopback  bool
}

type BackendConfig struct {
	signaling.BackendConfig

	Target  string
	Options BackendOptions
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) (err error) {
	c.BackendConfig = *cfg

	c.Target, c.Options, err = ParseURL(c.BackendConfig.URI.String())
	if err != nil {
		return fmt.Errorf("failed to parse multicast URL: %w", err)
	}

	return nil
}
