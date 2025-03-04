// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"fmt"

	"google.golang.org/grpc"

	"cunicu.li/cunicu/pkg/signaling"
)

type BackendConfig struct {
	signaling.BackendConfig

	Target  string
	Options []grpc.DialOption
}

func (c *BackendConfig) Parse(cfg *signaling.BackendConfig) (err error) {
	c.BackendConfig = *cfg

	c.Target, c.Options, err = ParseURL(c.BackendConfig.URI.String())
	if err != nil {
		return fmt.Errorf("failed to parse gRPC URL: %w", err)
	}

	return nil
}
