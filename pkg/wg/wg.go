// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package wg provides various helpers for WireGuard
package wg

import "errors"

const (
	SocketPath = "/var/run/wireguard"
	ConfigPath = "/etc/wireguard"

	DefaultPort = 51820

	TunnelOverhead = 80 // Byte
	DefaultMTU     = 1500 - TunnelOverhead
	MinimalMTU     = 1280 // Byte for minimal IPv6 MTU
)

var errNotSupported = errors.New("not supported")
